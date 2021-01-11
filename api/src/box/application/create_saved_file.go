package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/box/realtime"
)

// CreateSavedFileRequest ...
type CreateSavedFileRequest struct {
	EncryptedFileID   string `json:"encrypted_file_id"`
	IdentityID        string `json:"identity_id"`
	EncryptedMetadata string `json:"encrypted_metadata"`
	KeyFingerprint    string `json:"key_fingerprint"`
}

// BindAndValidate ...
func (req *CreateSavedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.EncryptedFileID, v.Required, is.UUIDv4),
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
		v.Field(&req.EncryptedMetadata, v.Required),
		v.Field(&req.KeyFingerprint, v.Required),
	)
}

// CreateSavedFile ...
func (app *BoxApplication) CreateSavedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateSavedFileRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merr.Unauthorized()
	}
	// check request identity consistency
	if req.IdentityID != access.IdentityID {
		return nil, merr.Forbidden().Add("identity_id", merr.DVForbidden)
	}

	// check identity has access to the original file
	hasAccess, err := events.HasAccessToFile(
		ctx, app.DB, app.RedConn,
		access.IdentityID, req.EncryptedFileID,
	)
	if err != nil {
		return nil, merr.From(err).Desc("checking file access")
	}
	if !hasAccess {
		return nil, merr.Forbidden()
	}

	// generate a new uuid as a saved file ID
	id, err := uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("generating saved file id")
	}
	// create the actual saved_file
	savedFile := files.SavedFile{
		ID:                id,
		IdentityID:        req.IdentityID,
		EncryptedFileID:   req.EncryptedFileID,
		EncryptedMetadata: req.EncryptedMetadata,
		KeyFingerprint:    req.KeyFingerprint,
	}
	if err := files.CreateSavedFile(ctx, app.DB, &savedFile); err != nil {
		return nil, merr.From(err).Desc("creating saved file")
	}

	// send websocket
	su := realtime.Update{
		Type: "file.saved",
		Object: struct {
			EncryptedFileID string `json:"encrypted_file_id"`
			IsSaved         bool   `json:"is_saved"`
		}{
			EncryptedFileID: savedFile.EncryptedFileID,
			IsSaved:         true,
		},
	}
	realtime.SendUpdate(ctx, app.RedConn, savedFile.IdentityID, &su)

	return savedFile, nil
}
