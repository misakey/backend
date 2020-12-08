package application

import (
	"context"
	"mime/multipart"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
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
		return merror.Transform(err).From(merror.OriBody)
	}
	req.identityID = eCtx.Param("id")
	file, err := eCtx.FormFile("encrypted_file")
	if err != nil {
		return merror.BadRequest().From(merror.OriBody).
			Detail("encrypted_file", merror.DVRequired).Describe(err.Error())
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
		return nil, merror.Unauthorized()
	}
	// user must have an account
	if acc.AccountID.IsZero() {
		return nil, merror.Forbidden().Describe("identity has no account")
	}
	if acc.IdentityID != req.identityID {
		return nil, merror.Forbidden()
	}

	// retrieve the raw []byte from the file
	encData, err := req.encFile.Open()
	if err != nil {
		return nil, merror.Internal().Describef("opening encrypted file: %v", err)
	}
	defer encData.Close()

	eFileID, err := uuid.NewString()
	if err != nil {
		return nil, merror.Transform(err).Describe("encrypted file id")
	}

	// create the encrypted file entity
	eFile := files.EncryptedFile{
		ID:   eFileID,
		Size: req.size,
	}
	if err := files.Create(ctx, app.DB, eFile); err != nil {
		return nil, merror.Transform(err).Describe("creating file")
	}

	// upload the encrypted data
	if err := files.Upload(ctx, app.filesRepo, eFileID, encData); err != nil {
		return nil, merror.Transform(err).Describe("uploading file")
	}

	sFileID, err := uuid.NewString()
	if err != nil {
		return nil, merror.Transform(err).Describe("saved file id")
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
		err = merror.Transform(err).Describe("inserting event in DB")
		if delErr := files.Delete(ctx, app.DB, app.filesRepo, eFileID); delErr != nil {
			return nil, merror.Transform(err).Describef("deleting file: %v", delErr)
		}
		return nil, err
	}

	return sFile, nil
}
