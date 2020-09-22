package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
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

func (bs *BoxApplication) CreateSavedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CreateSavedFileRequest)

	access := ajwt.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}
	// check identity
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	// get the boxes with this file
	events, err := events.FindByEncryptedFileID(ctx, bs.db, req.EncryptedFileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting events")
	}

	identityBoxEvents, err := boxes.LastSenderBoxEvents(ctx, bs.db, bs.redConn, access.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting identity boxes")
	}
	// creating an index to optimize next operation
	identityBoxesIndex := make(map[string]bool, len(identityBoxEvents))
	for _, boxEvent := range identityBoxEvents {
		identityBoxesIndex[boxEvent.BoxID] = true
	}

	// check that the identity has access to at least one box
	// containing this file
	accessForbidden := true
	for _, event := range events {
		if _, ok := identityBoxesIndex[event.BoxID]; ok {
			accessForbidden = false
			break
		}
	}
	if accessForbidden {
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
	if err := files.CreateSavedFile(ctx, bs.db, savedFile); err != nil {
		return nil, merror.Transform(err).Describe("creating saved file")
	}

	return savedFile, nil
}
