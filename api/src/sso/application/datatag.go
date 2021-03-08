package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/datatag"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/org"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// CreateDatatagCmd ...
type CreateDatatagCmd struct {
	Name           string `json:"name"`
	organizationID string
}

// BindAndValidate ...
func (cmd *CreateDatatagCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	cmd.organizationID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.organizationID, v.Required, is.UUIDv4),
		v.Field(&cmd.Name, v.Required),
	)
}

// CreateDatatag ...
func (sso *SSOService) CreateDatatag(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*CreateDatatagCmd)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// check the requester is admin of the organization
	if err := org.MustBeAdmin(ctx, sso.ssoDB, query.organizationID, acc.IdentityID); err != nil {
		return nil, merr.From(err).Desc("must be admin of the org")
	}

	// create the datatag
	id, err := uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("generating uuid")
	}
	datatag := &sqlboiler.Datatag{
		ID:             id,
		Name:           query.Name,
		OrganizationID: query.organizationID,
	}

	if err := datatag.Insert(ctx, sso.ssoDB, boil.Infer()); err != nil {
		return nil, merr.From(err).Desc("inserting datatag")
	}

	return datatag, nil
}

// ListDatatagsCmd ...
type ListDatatagsCmd struct {
	organizationID string
}

// BindAndValidate ...
func (cmd *ListDatatagsCmd) BindAndValidate(eCtx echo.Context) error {
	cmd.organizationID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.organizationID, v.Required, is.UUIDv4),
	)
}

// ListDatatags ...
func (sso *SSOService) ListDatatags(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ListDatatagsCmd)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// check the requester is admin of the org
	if err := org.MustBeAdmin(ctx, sso.ssoDB, query.organizationID, acc.IdentityID); err != nil {
		return nil, merr.From(err).Desc("must be admin of the org")
	}

	// list the datatags
	mods := []qm.QueryMod{
		sqlboiler.DatatagWhere.OrganizationID.EQ(query.organizationID),
	}
	datatags, err := sqlboiler.Datatags(mods...).All(ctx, sso.ssoDB)
	if err != nil {
		return nil, merr.From(err).Desc("list datatags")
	}
	if datatags == nil {
		return []*sqlboiler.Datatag{}, nil
	}

	return datatags, nil
}

// PatchDatatagCmd ...
type PatchDatatagCmd struct {
	organizationID string
	datatagID      string
	Name           string `json:"name"`
}

// BindAndValidate ...
func (cmd *PatchDatatagCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	cmd.organizationID = eCtx.Param("id")
	cmd.datatagID = eCtx.Param("did")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.Name, v.Required),
	)
}

// PatchDatatag ...
func (sso *SSOService) PatchDatatag(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*PatchDatatagCmd)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// check that the datatag exist
	datatag, err := datatag.Get(ctx, sso.ssoDB, query.datatagID)
	if err != nil {
		return nil, merr.From(err).Desc("getting datatag")
	}
	if query.organizationID != datatag.OrganizationID {
		return nil, merr.Forbidden()
	}

	// check that the user is an organization admin
	if err := org.MustBeAdmin(ctx, sso.ssoDB, query.organizationID, acc.IdentityID); err != nil {
		return nil, merr.From(err).Desc("must be admin of the org")
	}

	// edit the datatag
	datatag.Name = query.Name

	if _, err := datatag.Update(ctx, sso.ssoDB, boil.Whitelist(sqlboiler.DatatagColumns.Name)); err != nil {
		return nil, merr.From(err).Desc("editing datatag")
	}

	return nil, nil
}
