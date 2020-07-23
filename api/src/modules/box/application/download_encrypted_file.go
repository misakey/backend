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
	boxID  string
	fileID string
}

func (req *DownloadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.boxID = eCtx.Param("bid")
	req.fileID = eCtx.Param("eid")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.fileID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) DownloadEncryptedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*DownloadEncryptedFileRequest)
	acc := ajwt.GetAccesses(ctx)

	// if the box is closed, only the creator can download a file from it
	if err := boxes.MustBeCreatorIfClosed(ctx, bs.db, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	// check the box and the file does exist - represented by an event
	_, err := events.GetMsgFile(ctx, bs.db, req.boxID, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("finding msg.file event")
	}

	// download the file then render it
	data, err := files.Download(ctx, bs.filesRepo, req.boxID, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("downloading")
	}

	return data, nil
}
