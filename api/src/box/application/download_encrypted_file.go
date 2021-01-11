package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// DownloadEncryptedFileRequest ...
type DownloadEncryptedFileRequest struct {
	fileID string
}

// BindAndValidate ...
func (req *DownloadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	req.fileID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.fileID, v.Required, is.UUIDv4),
	)
}

// DownloadEncryptedFile ...
func (app *BoxApplication) DownloadEncryptedFile(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*DownloadEncryptedFileRequest)

	// check the file does exist
	_, err := files.Get(ctx, app.DB, req.fileID)
	if err != nil {
		return nil, merr.From(err).Desc("finding msg.file event")
	}

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	allowed, err := events.HasAccessOrHasSavedFile(ctx, app.DB, app.RedConn, acc.IdentityID, req.fileID)
	if err != nil {
		return nil, merr.From(err).Desc("checking access to file")
	}
	if !allowed {
		return nil, merr.Forbidden()
	}

	// download the file then render it
	readCloser, err := files.Download(ctx, app.filesRepo, req.fileID)
	if err != nil {
		return nil, merr.From(err).Desc("downloading")
	}
	return readCloser, nil
}
