package application

import (
	"context"
	"mime/multipart"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// UploadEncryptedFileRequest ...
type UploadEncryptedFileRequest struct {
	boxID string
	size  int64

	encFile *multipart.FileHeader

	MsgEncContent string `form:"msg_encrypted_content"`
	MsgPubKey     string `form:"msg_public_key"`
}

// BindAndValidate ...
func (req *UploadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	req.boxID = eCtx.Param("bid")
	file, err := eCtx.FormFile("encrypted_file")
	if err != nil {
		return merr.BadRequest().Ori(merr.OriBody).
			Add("encrypted_file", merr.DVRequired).Desc(err.Error())
	}
	req.encFile = file
	req.size = file.Size
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.MsgEncContent, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&req.MsgPubKey, v.Required),
		v.Field(&req.size, v.Required, v.Max(126*1024*1024).Error("the maximum file size is 126MB")), // @FIXME put the max file size in a configuration
	)
}

// UploadEncryptedFile ...
func (app *BoxApplication) UploadEncryptedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*UploadEncryptedFileRequest)

	// checking accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	// check box exists
	if err := events.MustBoxExists(ctx, app.DB, req.boxID); err != nil {
		return nil, merr.From(err).Desc("checking exist")
	}

	// retrieve the raw []byte from the file
	encData, err := req.encFile.Open()
	if err != nil {
		return nil, merr.Internal().Descf("opening encrypted file: %v", err)
	}
	defer encData.Close()

	// create the new msg file that will described the upload action
	e, fileID, err := events.NewMsgFile(ctx, req.boxID, acc.IdentityID, req.MsgEncContent, req.MsgPubKey)
	if err != nil {
		return nil, merr.From(err).Desc("creating msg file event")
	}

	// create the encrypted file entity
	eFile := files.EncryptedFile{
		ID:   fileID,
		Size: req.size,
	}
	if err := files.Create(ctx, app.DB, eFile); err != nil {
		return nil, merr.From(err).Desc("creating file")
	}

	// upload the encrypted data
	if err := files.Upload(ctx, app.filesRepo, fileID, encData); err != nil {
		return nil, merr.From(err).Desc("uploading file")
	}

	// persist the event in storage - on failure, we try to remove the uploaded file

	// set fileSize in content to compute boxUsedSpace after event is persisted
	metadata := events.MetadataForUsedSpaceHandler{
		NewEventSize: req.size,
	}

	eReq := CreateEventRequest{
		boxID:               e.BoxID,
		Type:                e.Type,
		Content:             e.JSONContent,
		MetadataForHandlers: metadata,
	}
	view, err := app.CreateEvent(ctx, &eReq)
	if err != nil {
		err = merr.From(err).Desc("inserting event in DB")
		if delErr := files.Delete(ctx, app.DB, app.filesRepo, fileID); delErr != nil {
			return nil, merr.From(err).Descf("deleting file: %v", delErr)
		}
		return nil, err
	}

	return view, nil
}
