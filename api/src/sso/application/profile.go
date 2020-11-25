package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

//
// Configure profile
//

type ConfigProfileCmd struct {
	identityID string
	ShareEmail *bool `json:"email"`
}

func (cmd *ConfigProfileCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.ShareEmail, v.NotNil),
	)
}

func (sso *SSOService) SetProfileConfig(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ConfigProfileCmd)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	if *query.ShareEmail {
		err = identity.ProfileConfigShare(ctx, tr, query.identityID, string(identity.EmailIdentifier))
	} else {
		err = identity.ProfileConfigUnshare(ctx, tr, query.identityID, string(identity.EmailIdentifier))
	}
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

//
// Get current profile configuration
//

type ConfigProfileQuery struct {
	identityID string
}

func (query *ConfigProfileQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")
	return v.ValidateStruct(query, v.Field(&query.identityID, v.Required, is.UUIDv4))
}

func (sso *SSOService) GetProfileConfig(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ConfigProfileQuery)

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merror.Forbidden()
	}

	return identity.ProfileConfigGet(ctx, sso.sqlDB, query.identityID)
}

//
// Get profile
//

type ProfileQuery struct {
	identityID string
}

func (query *ProfileQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")

	return v.ValidateStruct(query, v.Field(&query.identityID, v.Required, is.UUIDv4))
}

func (sso *SSOService) GetProfile(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ProfileQuery)
	return identity.ProfileGet(ctx, sso.sqlDB, query.identityID)
}
