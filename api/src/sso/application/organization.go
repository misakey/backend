package application

import (
	"context"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/org"
)

// OrgView ...
type OrgView struct {
	org.Org
	CurrentIdentityRole null.String `json:"current_identity_role"`
}

// OrgCreateCmd ...
type OrgCreateCmd struct {
	Name string `json:"name"`
}

// BindAndValidate ...
func (cmd *OrgCreateCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	return v.ValidateStruct(cmd,
		v.Field(&cmd.Name, v.Required),
	)
}

// CreateOrg ...
func (sso *SSOService) CreateOrg(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*OrgCreateCmd)
	view := OrgView{}
	var err error

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return view, merr.Forbidden()
	}

	view.Org, err = org.Create(ctx, sso.sqlDB, acc.IdentityID, query.Name)
	if err != nil {
		return nil, err
	}
	// since the org has just been created, the current identity is the admin
	view.CurrentIdentityRole = null.StringFrom("admin")
	return view, nil
}

// OrgListQuery ...
type OrgListQuery struct {
	identityID string
}

// BindAndValidate ...
func (query *OrgListQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")
	return v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	)
}

// ListIdentityOrgs ...
func (sso *SSOService) ListIdentityOrgs(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*OrgListQuery)
	views := []OrgView{}
	var err error

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return views, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return views, merr.Forbidden()
	}

	// the list is composed by:
	// - the self-org
	// - organization where the user is a member of
	// - organization where the user has a role

	// 1. add the self-org
	curIdentity, err := identity.Get(ctx, sso.sqlDB, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identity")
	}
	// no identity role is added for that org
	selfOrg := OrgView{
		Org: org.Org{
			ID:        sso.selfOrgID,
			Name:      curIdentity.DisplayName,
			CreatorID: curIdentity.ID,
			CreatedAt: time.Now(),
		},
	}
	views = append(views, selfOrg)

	// 2. get org ids where the identity is a member
	memberOrgIDs, err := org.GetIDsForIdentity(ctx, sso.redConn, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("listing member org ids")
	}

	// 3. get orgs where user has a role:
	// - only orgs that they have created
	// query both member orgs and created orgs
	orgs, err := org.ListByIDsOrCreatorID(ctx, sso.sqlDB, memberOrgIDs, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("listing created orgs")
	}
	for _, o := range orgs {
		view := OrgView{Org: o}
		// creators are admins
		if o.CreatorID == acc.IdentityID {
			view.CurrentIdentityRole = null.StringFrom("admin")
		}
		views = append(views, view)
	}
	return views, nil
}
