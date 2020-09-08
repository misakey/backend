package application

import (
	"context"
	"mime/multipart"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type UploadEncryptedFileRequest struct {
	boxID string
	size  int64

	encFile *multipart.FileHeader

	MsgEncContent string `form:"msg_encrypted_content"`
	MsgPubKey     string `form:"msg_public_key"`
}

func (req *UploadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("bid")
	file, err := eCtx.FormFile("encrypted_file")
	if err != nil {
		return merror.BadRequest().From(merror.OriBody).
			Detail("encrypted_file", merror.DVRequired).Describe(err.Error())
	}
	req.encFile = file
	req.size = file.Size
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.MsgEncContent, v.Required, is.Base64),
		v.Field(&req.MsgPubKey, v.Required),
		v.Field(&req.size, v.Required, v.Max(8*1024*1000).Error("the maximum file size is 8MB")),
	)
}

func (bs *BoxApplication) UploadEncryptedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*UploadEncryptedFileRequest)

	// checking accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, bs.db, bs.identities, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	// upload files works only on open boxes
	if err := events.MustBoxBeOpen(ctx, bs.db, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("checking open")
	}

	// retrieve the raw []byte from the file
	encData, err := req.encFile.Open()
	if err != nil {
		return nil, merror.Internal().Describef("opening encrypted file: %v", err)
	}
	defer encData.Close()

	// create the new msg file that will described the upload action
	e, fileID, err := events.NewMsgFile(ctx, req.boxID, acc.IdentityID, req.MsgEncContent, req.MsgPubKey)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating msg file event")
	}

	// create the encrypted file entity
	eFile := files.EncryptedFile{
		ID:   fileID,
		Size: req.size,
	}
	if err := files.Create(ctx, bs.db, eFile); err != nil {
		return nil, merror.Transform(err).Describe("creating file")
	}

	// upload the encrypted data
	if err := files.Upload(ctx, bs.filesRepo, fileID, encData); err != nil {
		return nil, merror.Transform(err).Describe("uploading file")
	}

	// persist the event in storage - on failure, we try to remove the uploaded file
	eReq := CreateEventRequest{
		boxID:   e.BoxID,
		Type:    e.Type,
		Content: e.JSONContent,
	}
	view, err := bs.CreateEvent(ctx, &eReq)
	if err != nil {
		err = merror.Transform(err).Describe("inserting event in DB")
		if delErr := files.Delete(ctx, bs.db, bs.filesRepo, fileID); delErr != nil {
			return nil, merror.Transform(err).Describef("deleting file: %v", delErr)
		}
		return nil, err
	}

	return view, nil
}
