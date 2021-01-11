package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/box/realtime"
)

// DeleteSavedFileRequest ...
type DeleteSavedFileRequest struct {
	ID string
}

// BindAndValidate ...
func (req *DeleteSavedFileRequest) BindAndValidate(eCtx echo.Context) error {
	req.ID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.ID, v.Required, is.UUIDv4),
	)
}

// DeleteSavedFile ...
func (app *BoxApplication) DeleteSavedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*DeleteSavedFileRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merr.Unauthorized()
	}

	// get saved file
	savedFile, err := files.GetSavedFile(ctx, app.DB, req.ID)
	if err != nil {
		return nil, merr.From(err).Desc("getting saved file")
	}
	if savedFile.IdentityID != access.IdentityID {
		return nil, merr.Forbidden().Add("id", merr.DVForbidden)
	}

	if err := files.DeleteSavedFile(ctx, app.DB, req.ID); err != nil {
		return nil, err
	}

	// delete stored file if orphan
	isOrphan, err := events.IsFileOrphan(ctx, app.DB, savedFile.EncryptedFileID)
	if err != nil {
		return nil, merr.From(err).Desc("deleting stored file")
	}
	if isOrphan {
		if err := files.Delete(ctx, app.DB, app.filesRepo, savedFile.EncryptedFileID); err != nil {
			return nil, merr.From(err).Desc("deleting stored file")
		}
	}

	// send websocket
	su := realtime.Update{
		Type: "file.saved",
		Object: struct {
			EncryptedFileID string `json:"encrypted_file_id"`
			IsSaved         bool   `json:"is_saved"`
		}{
			EncryptedFileID: savedFile.EncryptedFileID,
			IsSaved:         false,
		},
	}
	realtime.SendUpdate(ctx, app.RedConn, savedFile.IdentityID, &su)

	return nil, nil
}
