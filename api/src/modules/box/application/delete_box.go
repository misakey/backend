package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshares"
)

type DeleteBoxRequest struct {
	boxID string

	UserConfirmation string `json:"user_confirmation"`
}

func (req *DeleteBoxRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		// verify the user has entered a valid confirmation key
		v.Field(&req.UserConfirmation, v.Required, v.In("delete", "supprimer").Error("must be delete|supprimer")),
	)
}

func (bs *BoxApplication) DeleteBox(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*DeleteBoxRequest)

	acc := ajwt.GetAccesses(ctx)

	// 1. verify the deletion sender is an admin of the box
	if err := boxes.MustBeCreator(ctx, bs.db, req.boxID, acc.IdentityID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// init a transaction to ensure all entities are removed
	tr, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initing transaction")
	}
	defer atomic.SQLRollback(ctx, tr, err)

	// 2. Delete all the events
	if err := events.DeleteAllForBox(ctx, tr, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("deleting events")
	}

	// 3. Delete the key shares
	if err := keyshares.EmptyAll(ctx, tr, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("emptying keyshares")
	}

	// 4. Delete orphan files
	boxFileIDs, err := events.ListFilesID(ctx, bs.db, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting files")
	}
	for _, fileID := range boxFileIDs {
		isOrphan, err := files.IsOrphan(ctx, bs.db, fileID)
		if err != nil {
			return nil, merror.Transform(err).Describe("checking file is orphan")
		}
		if isOrphan {
			if err := files.Delete(ctx, bs.db, bs.filesRepo, fileID); err != nil {
				return nil, merror.Transform(err).Describe("deleting stored file")
			}
		}
	}

	return nil, tr.Commit()
}
