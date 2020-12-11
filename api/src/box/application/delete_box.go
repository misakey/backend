package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/realtime"
)

// DeleteBoxRequest ...
type DeleteBoxRequest struct {
	boxID string

	UserConfirmation string `json:"user_confirmation"`
}

// BindAndValidate ...
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

// DeleteBox ...
func (app *BoxApplication) DeleteBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*DeleteBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// verify the deletion sender is an admin of the box
	if err := events.MustBeAdmin(ctx, app.DB, req.boxID, acc.IdentityID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// get box files before deleting events
	boxFileIDs, err := events.ListFilesID(ctx, app.DB, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting files")
	}

	// get public key before deleting events
	boxPublicKey, err := events.GetBoxPublicKey(ctx, app.DB, req.boxID)
	if err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not get publicKey for %s", req.boxID)
	}

	// get box members (to notify them)
	memberIDs, err := events.ListBoxMemberIDs(ctx, app.DB, app.RedConn, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting members list")
	}

	// init a transaction to ensure all entities are removed
	tr, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initing transaction")
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	if err := events.CleanBox(ctx, tr, req.boxID); err != nil {
		return nil, err
	}

	if err := events.DeleteOrphanFiles(ctx, tr, app.filesRepo, boxFileIDs); err != nil {
		return nil, err
	}

	// run db operations
	if cErr := tr.Commit(); cErr != nil {
		return nil, merror.Transform(cErr).Describe("committing transaction")
	}

	// send delete events to websockets
	bu := realtime.Update{
		Type: "box.delete",
		Object: struct {
			BoxID     string `json:"id"`
			SenderID  string `json:"sender_id"`
			PublicKey string `json:"public_key"`
		}{
			BoxID:     req.boxID,
			SenderID:  acc.IdentityID,
			PublicKey: boxPublicKey,
		},
	}
	for _, memberID := range memberIDs {
		realtime.SendUpdate(ctx, app.RedConn, memberID, &bu)
	}

	// clean up some redis keys
	if err := cache.CleanBoxCache(ctx, app.RedConn, req.boxID); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("cleaning box %s cache", req.boxID)
	}

	// invalidate cache for members
	// to avoid having this box in user lists
	for _, memberID := range memberIDs {
		if err := cache.CleanBoxesListCache(ctx, app.RedConn, memberID); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("cleaning boxes list for %s cache", memberID)
		}
	}

	return nil, nil

}
