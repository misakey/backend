package application

import (
	"context"
	"mime/multipart"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// UploadSavedFileRequest ...
type UploadSavedFileRequest struct {
	identityID string
	size       int64

	encFile *multipart.FileHeader

	EncryptedMetadata string `form:"encrypted_metadata"`
	KeyFingerprint    string `form:"key_fingerprint"`
}

// BindAndValidate ...
func (req *UploadSavedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	req.identityID = eCtx.Param("id")
	file, err := eCtx.FormFile("encrypted_file")
	if err != nil {
		return merr.BadRequest().Ori(merr.OriBody).
			Add("encrypted_file", merr.DVRequired).Desc(err.Error())
	}
	req.encFile = file
	req.size = file.Size
	return v.ValidateStruct(req,
		v.Field(&req.identityID, v.Required, is.UUIDv4),
		v.Field(&req.EncryptedMetadata, v.Required),
		v.Field(&req.KeyFingerprint, v.Required),
		v.Field(&req.size, v.Required, v.Max(126*1024*1024).Error("the maximum file size is 126MB")), // @FIXME put the max file size in a configuration
	)
}

// UploadSavedFile ...
func (app *BoxApplication) UploadSavedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*UploadSavedFileRequest)

	// checking accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	// user must have an account
	if acc.AccountID.IsZero() {
		return nil, merr.Forbidden().Desc("identity has no account")
	}
	if acc.IdentityID != req.identityID {
		return nil, merr.Forbidden()
	}

	// retrieve the raw []byte from the file
	encData, err := req.encFile.Open()
	if err != nil {
		return nil, merr.Internal().Descf("opening encrypted file: %v", err)
	}
	defer encData.Close()

	eFileID, err := uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("encrypted file id")
	}

	// create the encrypted file entity
	eFile := files.EncryptedFile{
		ID:   eFileID,
		Size: req.size,
	}
	if err := files.Create(ctx, app.DB, eFile); err != nil {
		return nil, merr.From(err).Desc("creating file")
	}

	// upload the encrypted data
	if err := files.Upload(ctx, app.filesRepo, eFileID, encData); err != nil {
		return nil, merr.From(err).Desc("uploading file")
	}

	sFileID, err := uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("saved file id")
	}

	// persist saved file
	// in case of failure, we try to revert the file upload
	sFile := files.SavedFile{
		ID:                sFileID,
		IdentityID:        req.identityID,
		EncryptedFileID:   eFileID,
		EncryptedMetadata: req.EncryptedMetadata,
		KeyFingerprint:    req.KeyFingerprint,
	}

	if err := files.CreateSavedFile(ctx, app.DB, &sFile); err != nil {
		err = merr.From(err).Desc("inserting event in DB")
		if delErr := files.Delete(ctx, app.DB, app.filesRepo, eFileID); delErr != nil {
			return nil, merr.From(err).Descf("deleting file: %v", delErr)
		}
		return nil, err
	}

	return sFile, nil
}
