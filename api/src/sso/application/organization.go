package application

import (
	"context"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mrand"
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

	view.Org, err = org.Create(ctx, sso.ssoDB, acc.IdentityID, query.Name)
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
	curIdentity, err := identity.Get(ctx, sso.ssoDB, acc.IdentityID)
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
	memberOrgIDs, err := org.GetIDsForIdentity(ctx, sso.boxDB, sso.redConn, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("listing member org ids")
	}

	// 3. get orgs where user has a role:
	// - only orgs that they have created
	// query both member orgs and created orgs
	orgs, err := org.ListByIDsOrCreatorID(ctx, sso.ssoDB, memberOrgIDs, acc.IdentityID)
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

// GetOrgPublicRequest ...
type GetOrgPublicRequest struct {
	orgID string
}

// BindAndValidate ...
func (req *GetOrgPublicRequest) BindAndValidate(eCtx echo.Context) error {
	req.orgID = eCtx.Param("id")
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.orgID, v.Required, is.UUIDv4),
	)
}

// PublicOrgView ...
type PublicOrgView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}

// GetOrgPublic returns public data.
// No access check performed
func (sso *SSOService) GetOrgPublic(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetOrgPublicRequest)

	// get org title
	organization, err := org.GetOrg(ctx, sso.ssoDB, req.orgID)
	if err != nil {
		return nil, err
	}

	view := PublicOrgView{
		ID:      organization.ID,
		Name:    organization.Name,
		LogoURL: organization.LogoURL.String,
	}
	return view, nil
}

type GenerateSecretCmd struct {
	orgID string
}

func (cmd *GenerateSecretCmd) BindAndValidate(eCtx echo.Context) error {
	cmd.orgID = eCtx.Param("id")
	return v.ValidateStruct(cmd,
		v.Field(&cmd.orgID, v.Required, is.UUIDv4),
	)
}

type SecretView struct {
	Secret string `json:"secret"`
}

// GenerateSecret for the received organization id. Requires admin accesses.
// - create the hydra client if not existing yet
// - create an identity corresponding to the org if not existing yet
// - update the hydra secret and return it in json
func (sso *SSOService) GenerateSecret(ctx context.Context, genReq request.Request) (interface{}, error) {
	cmd := genReq.(*GenerateSecretCmd)

	// 1. check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if err := org.MustBeAdmin(ctx, sso.ssoDB, cmd.orgID, acc.IdentityID); err != nil {
		return nil, merr.From(err).Desc("must be admin of the org")
	}

	// generate the new secret - size 32 for no concrete reason
	secret, err := mrand.Base64String(32)
	if err != nil {
		return nil, merr.From(err).Desc("generating new secret")
	}

	err = sso.authFlowService.UpdateClientSecret(ctx, cmd.orgID, secret)
	if err != nil {
		return "", err
	}

	// since organization might perform some api requests after generating a secret,
	// ensure there is a identity corresponding to the org
	_, err = identity.GetByIdentifier(ctx, sso.ssoDB, cmd.orgID, identity.IdentifierKindOrgID)
	if err != nil {
		// create the identity on not found
		if merr.IsANotFound(err) {
			orga, err := org.GetOrg(ctx, sso.ssoDB, cmd.orgID)
			if err != nil {
				return nil, merr.From(err).Desc("getting org")
			}
			orgIdentity := identity.Identity{
				ID:              cmd.orgID,
				IdentifierValue: cmd.orgID,
				IdentifierKind:  identity.IdentifierKindOrgID,
				DisplayName:     orga.Name,
			}
			if err := identity.Create(ctx, sso.ssoDB, sso.redConn, &orgIdentity); err != nil {
				return nil, merr.From(err).Desc("creating org identity")
			}
		} else { // otherwise return the err
			return nil, merr.From(err).Desc("getting org identity")
		}
	}
	// bind and return view
	return SecretView{secret}, nil
}
