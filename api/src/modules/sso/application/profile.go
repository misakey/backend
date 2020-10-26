package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
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

	if *query.ShareEmail {
		return nil, sso.identityService.ProfileConfigShare(ctx, query.identityID, string(domain.EmailIdentifier))
	} else {
		return nil, sso.identityService.ProfileConfigUnshare(ctx, query.identityID, string(domain.EmailIdentifier))
	}
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

	return sso.identityService.ProfileConfigGet(ctx, query.identityID)
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
	return sso.identityService.ProfileGet(ctx, query.identityID)
}
