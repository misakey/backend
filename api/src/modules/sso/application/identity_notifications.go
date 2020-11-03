package application

import (
	"context"
	"strconv"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

type IdentityNotifCountQuery struct {
	identityID string
}

func (query *IdentityNotifCountQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")

	if err := v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating identity notification count query")
	}
	return nil
}

func (sso *SSOService) CountIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityNotifCountQuery)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != query.identityID {
		return -1, merror.Forbidden()
	}

	return sso.identityService.NotificationCount(ctx, sso.sqlDB, query.identityID)
}

type IdentityNotifListQuery struct {
	identityID string

	Offset null.Int `query:"offset"`
	Limit  null.Int `query:"limit"`
}

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

func (sso *SSOService) ListIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityNotifListQuery)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != query.identityID {
		return nil, merror.Forbidden()
	}

	// list notifs
	notifs, err := sso.identityService.NotificationList(ctx, sso.sqlDB, query.identityID, query.Offset, query.Limit)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing identity notification")
	}
	return notifs, nil
}

type IdentityNotifAckCmd struct {
	identityID string
	notifIDs   []int

	StrNotifIDs string `query:"ids"`
}

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

func (sso *SSOService) AckIdentityNotification(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*IdentityNotifAckCmd)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}
	return nil, sso.identityService.NotificationAck(ctx, sso.sqlDB, cmd.identityID, cmd.notifIDs)
}
