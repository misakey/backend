package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type DownloadEncryptedFileRequest struct {
	fileID string
}

func (req *DownloadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	req.fileID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.fileID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) DownloadEncryptedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*DownloadEncryptedFileRequest)
	acc := ajwt.GetAccesses(ctx)

	// check the file does exist
	_, err := files.Get(ctx, bs.db, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("finding msg.file event")
	}

	isAllowed := false

	// get all entities linked to the file
	linkedEvents, err := events.FindByEncryptedFileID(ctx, bs.db, req.fileID)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return nil, err
	}
	linkedSavedFiles, err := files.ListSavedFilesByFileID(ctx, bs.db, req.fileID)
	if err != nil {
		return nil, err
	}

	for _, event := range linkedEvents {
		closedErr := boxes.MustBeCreatorIfClosed(ctx, bs.db, event.BoxID, acc.IdentityID)
		actorErr := boxes.MustBeActor(ctx, bs.db, event.BoxID, acc.IdentityID)
		if closedErr == nil && actorErr == nil {
			isAllowed = true
			break
		}

	}

	if !isAllowed {
		for _, savedFile := range linkedSavedFiles {
			if savedFile.IdentityID == acc.IdentityID {
				isAllowed = true
				break
			}
		}
	}

	if !isAllowed {
		return nil, merror.Forbidden()
	}

	// download the file then render it
	data, err := files.Download(ctx, bs.filesRepo, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("downloading")
	}

	return data, nil
}
