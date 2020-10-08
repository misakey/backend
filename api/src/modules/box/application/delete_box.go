package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshares"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/quota"
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

func (bs *BoxApplication) DeleteBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*DeleteBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// 1. verify the deletion sender is an admin of the box
	if err := events.MustBeAdmin(ctx, bs.DB, req.boxID, acc.IdentityID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// get box files before deleting events
	boxFileIDs, err := events.ListFilesID(ctx, bs.DB, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting files")
	}

	// get box members (to notify them)
	memberIDs, err := events.ListBoxMemberIDs(ctx, bs.DB, bs.RedConn, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting members list")
	}

	// init a transaction to ensure all entities are removed
	tr, err := bs.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initing transaction")
	}
	defer atomic.SQLRollback(ctx, tr, err)

	// 2. Delete all the events
	if err := events.DeleteAllForBox(ctx, tr, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("deleting events")
	}

	// 3. Get public key
	createEvent, err := events.GetCreateEvent(ctx, bs.DB, req.boxID)
	if err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not get publicKey for %s", req.boxID)
	}
	creationContent := events.CreationContent{}
	if err = createEvent.JSONContent.Unmarshal(&creationContent); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not get publicKey for %s", req.boxID)
	}

	// 4. Delete the key shares
	if err := keyshares.EmptyAll(ctx, tr, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("emptying keyshares")
	}

	// 5. Delete the box used space
	if err := quota.DeleteBoxUsedSpace(ctx, tr, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("emptying box used space")
	}

	// run db operations
	if err := tr.Commit(); err != nil {
		return nil, err
	}

	// 6. Delete orphan files
	for _, fileID := range boxFileIDs {
		// we need to check the existency of fileID
		// since it is set to "" when msg.delete is called on the msg.file
		// TODO: clean this up
		if fileID != "" {
			isOrphan, err := events.IsFileOrphan(ctx, bs.DB, fileID)
			if err != nil {
				return nil, merror.Transform(err).Describe("checking file is orphan")
			}
			if isOrphan {
				if err := files.Delete(ctx, bs.DB, bs.filesRepo, fileID); err != nil {
					return nil, merror.Transform(err).Describe("deleting stored file")
				}
			}
		}
	}

	// 7. Send event to websockets
	events.SendDeleteBox(ctx, bs.RedConn, req.boxID, acc.IdentityID, memberIDs, creationContent.PublicKey)

	return nil, nil

}
