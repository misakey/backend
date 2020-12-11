package application

import (
	"context"
	"strconv"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// IdentityNotifCountQuery ...
type IdentityNotifCountQuery struct {
	identityID string
}

// BindAndValidate ...
func (query *IdentityNotifCountQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")

	if err := v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating identity notification count query")
	}
	return nil
}

// CountIdentityNotification ...
func (sso *SSOService) CountIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityNotifCountQuery)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != query.identityID {
		return -1, merror.Forbidden()
	}

	return identity.NotificationCount(ctx, sso.sqlDB, query.identityID)
}

// IdentityNotifListQuery ...
type IdentityNotifListQuery struct {
	identityID string

	Offset null.Int `query:"offset"`
	Limit  null.Int `query:"limit"`
}

// BindAndValidate ...
func (query *IdentityNotifListQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	query.identityID = eCtx.Param("id")

	if err := v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating identity notification list query")
	}
	return nil
}

// ListIdentityNotification ...
func (sso *SSOService) ListIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityNotifListQuery)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != query.identityID {
		return nil, merror.Forbidden()
	}

	// list notifs
	notifs, err := identity.NotificationList(ctx, sso.sqlDB, query.identityID, query.Offset, query.Limit)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing identity notification")
	}
	return notifs, nil
}

// IdentityNotifAckCmd ...
type IdentityNotifAckCmd struct {
	identityID string
	notifIDs   []int

	StrNotifIDs string `query:"ids"`
}

// BindAndValidate ...
func (cmd *IdentityNotifAckCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.Transform(err).From(merror.OriQuery)
	}

	strSliceIDs := strings.Split(cmd.StrNotifIDs, ",")
	for _, strID := range strSliceIDs {
		id, err := strconv.ParseUint(strID, 10, 32)
		if err != nil {
			return merror.Transform(err).From(merror.OriQuery).Detail("ids", merror.DVMalformed)
		}
		cmd.notifIDs = append(cmd.notifIDs, int(id))
	}

	cmd.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating identity notification acknowledged cmd")
	}
	return nil
}

// AckIdentityNotification ...
func (sso *SSOService) AckIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*IdentityNotifAckCmd)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = identity.NotificationAck(ctx, tr, cmd.identityID, cmd.notifIDs)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}
