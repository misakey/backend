package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type CreateSavedFileRequest struct {
	EncryptedFileID   string `json:"encrypted_file_id"`
	IdentityID        string `json:"identity_id"`
	EncryptedMetadata string `json:"encrypted_metadata"`
	KeyFingerprint    string `json:"key_fingerprint"`
}

func (req *CreateSavedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.EncryptedFileID, v.Required, is.UUIDv4),
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
		v.Field(&req.EncryptedMetadata, v.Required),
		v.Field(&req.KeyFingerprint, v.Required),
	)
}

func (app *BoxApplication) CreateSavedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateSavedFileRequest)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}
	// check request identity consistency
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	// check identity has access to the original file
	hasAccess, err := events.HasAccessToFile(
		ctx, app.DB, app.RedConn, identityMapper,
		access.IdentityID, req.EncryptedFileID,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("checking file access")
	}
	if !hasAccess {
		return nil, merror.Forbidden()
	}

	// generate a new uuid as a saved file ID
	id, err := uuid.NewString()
	if err != nil {
		return nil, merror.Transform(err).Describe("generating saved file id")
	}
	// create the actual saved_file
	savedFile := files.SavedFile{
		ID:                id,
		IdentityID:        req.IdentityID,
		EncryptedFileID:   req.EncryptedFileID,
		EncryptedMetadata: req.EncryptedMetadata,
		KeyFingerprint:    req.KeyFingerprint,
	}
	if err := files.CreateSavedFile(ctx, app.DB, savedFile); err != nil {
		return nil, merror.Transform(err).Describe("creating saved file")
	}

	return savedFile, nil
}
