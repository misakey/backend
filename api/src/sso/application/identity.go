package application

import (
	"context"
	"io"
	"path/filepath"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	_ "gitlab.misakey.dev/misakey/backend/api/src/sso/gamification"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

type IdentityQuery struct {
	identityID string
}

func (query *IdentityQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating identity query")
	}
	return nil
}

type IdentityView struct {
	identity.Identity
	HasAccount bool `json:"has_account"`
}

func (sso *SSOService) GetIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityQuery)
	view := IdentityView{}
	var err error

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden()
	}

	if acc.IdentityID != query.identityID {
		return view, merror.Forbidden()
	}

	// set the view on identity retrieval
	view.Identity, err = identity.Get(ctx, sso.sqlDB, query.identityID)
	if err != nil {
		return view, err
	}

	// always tell the client either the identity has a linked account or not
	view.HasAccount = view.Identity.AccountID.Valid
	// removes account id information if the end user is not connected to the account but only on the identity
	if !acc.AccountConnected() {
		view.Identity.AccountID = null.String{}
	}
	return view, err
}

// PartialUpdateIdentityCmd
type PartialUpdateIdentityCmd struct {
	identityID    string
	DisplayName   string      `json:"display_name"`
	Notifications string      `json:"notifications"`
	Color         null.String `json:"color"`
	Pubkey        null.String `json:"pubkey"`
}

// Validate the IdentityAuthableCmd
func (cmd *PartialUpdateIdentityCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Notifications, v.In("minimal", "moderate", "frequent")),
		v.Field(&cmd.DisplayName, v.Length(3, 254)),
		v.Field(&cmd.Color, v.Length(7, 7)),
	); err != nil {
		return merror.Transform(err).Describe("validating identity patch")
	}
	return nil
}

// PartialUpdateIdentity to change its display name or avatar
func (sso *SSOService) PartialUpdateIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*PartialUpdateIdentityCmd)
	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	var curIdentity identity.Identity
	curIdentity, err = identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	if cmd.DisplayName != "" {
		curIdentity.DisplayName = cmd.DisplayName
	}

	if cmd.Notifications != "" {
		curIdentity.Notifications = cmd.Notifications
	}

	if cmd.Color.Valid {
		curIdentity.Color = cmd.Color
	}

	if cmd.Pubkey.Valid {
		curIdentity.Pubkey = cmd.Pubkey
	}

	err = identity.Update(ctx, tr, &curIdentity)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

// UploadAvatarCmd
type UploadAvatarCmd struct {
	identityID string
	Data       io.Reader
	Extension  string
}

func (cmd *UploadAvatarCmd) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")

	file, err := eCtx.FormFile("avatar")
	if err != nil {
		return merror.BadRequest().From(merror.OriBody).Detail("avatar", merror.DVRequired).Describe(err.Error())
	}
	if file.Size >= 3*1024*1024 {
		return merror.BadRequest().From(merror.OriBody).Detail("size", merror.DVInvalid).Describe("size must be < 3 mo")
	}

	data, err := file.Open()
	if err != nil {
		return err
	}

	cmd.Data = data
	cmd.Extension = filepath.Ext(file.Filename)

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Extension, v.Required),
	)
}

func (sso *SSOService) UploadAvatar(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*UploadAvatarCmd)
	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	// get avatar's corresponding user
	var existingIdentity identity.Identity
	existingIdentity, err = identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	// first remove existing avatar if it does exist
	if !existingIdentity.AvatarURL.IsZero() {
		avatarToDel := &identity.AvatarFile{
			Filename: filepath.Base(existingIdentity.AvatarURL.String),
		}
		err = sso.identityService.DeleteAvatar(ctx, avatarToDel)
		if err != nil {
			return nil, err
		}
	}

	avatar := identity.AvatarFile{}
	// generate an UUID to use as a filename
	var filename uuid.UUID
	filename, err = uuid.NewRandom()
	if err != nil {
		return nil, merror.Transform(err).Describe("could not generate uuid v4")
	}
	avatar.Filename = filename.String() + avatar.Extension
	avatar.Extension = cmd.Extension
	avatar.Data = cmd.Data

	// upload the avatar to storage
	var url string
	url, err = sso.identityService.UploadAvatar(ctx, &avatar)
	if err != nil {
		return nil, err
	}

	// then save into user account the new avatar uri
	existingIdentity.AvatarURL = null.StringFrom(url)
	err = identity.Update(ctx, tr, &existingIdentity)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

// DeleteAvatarCmd
type DeleteAvatarCmd struct {
	identityID string
}

func (cmd *DeleteAvatarCmd) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")
	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// DeleteAvatar for a given identity
func (sso *SSOService) DeleteAvatar(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*DeleteAvatarCmd)
	avatar := identity.AvatarFile{}

	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	// get identity
	var curIdentity identity.Identity
	curIdentity, err = identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	// check that the identity has an avatar to delet
	if curIdentity.AvatarURL.IsZero() {
		err = merror.Conflict().Describe("avatar is not set").Detail("identity_id", merror.DVConflict)
		return nil, err
	}

	// delete avatar
	avatar.Filename = filepath.Base(curIdentity.AvatarURL.String)
	err = sso.identityService.DeleteAvatar(ctx, &avatar)
	if err != nil {
		return nil, err
	}

	// update identity with empty avatar url field
	curIdentity.AvatarURL = null.NewString("", false)
	err = identity.Update(ctx, tr, &curIdentity)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}

// AttachCouponCmd
type AttachCouponCmd struct {
	identityID string
	Value      string `json:"value"`
}

func (cmd *AttachCouponCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Value, v.Required),
	)
}

// AttachCoupon to a given identity
func (sso *SSOService) AttachCoupon(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*AttachCouponCmd)
	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merror.Forbidden()
	}

	// start transaction since write actions will be performed across entities
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, err)

	// get identity
	var curIdentity identity.Identity
	curIdentity, err = identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	// Check the coupon hasn't already been applied
	if curIdentity.Level >= 20 {
		err = merror.Conflict().Describe("coupon already applied")
		return nil, err
	}

	// NOTE: there is no valid coupon nowadays
	err = merror.BadRequest().Detail("value", merror.DVInvalid).Detail("invalid_value", cmd.Value).Describe("invalid coupon")
	return nil, err

	// // 3. Update the identity
	// curIdentity.Level = 20
	// err = identity.Update(ctx, tr, &curIdentity)
	// if err != nil {
	// 	return nil, merror.Transform(err).Describe("updating identity")
	// }

	// // 2. Create the used_coupon
	// usedCoupon := gamification.UsedCoupon{
	// 	IdentityID: cmd.identityID,
	// 	Value:      cmd.Value,
	// }

	// err = gamification.UseCoupon(ctx, tr, usedCoupon)
	// if err != nil {
	// 	return nil, merror.Transform(err).Describe("creating coupon")
	// }
	// return nil, tr.Commit()
}

type IdentityPubkeyByIdentifierQuery struct {
	IdentifierValue string `query:"identifier_value"`
}

func (query *IdentityPubkeyByIdentifierQuery) BindAndValidate(eCtx echo.Context) error {
	err := eCtx.Bind(query)
	if err != nil {
		return merror.BadRequest().Describe(err.Error())
	}
	return v.ValidateStruct(query,
		v.Field(&query.IdentifierValue, v.Required),
	)
}

func (sso *SSOService) GetIdentityPubkeyByIdentifier(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityPubkeyByIdentifierQuery)

	// must only be authenticated
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}

	identities, err := identity.ListByIdentifier(ctx, sso.sqlDB,
		identity.Identifier{
			Value: query.IdentifierValue,
			Kind:  identity.EmailIdentifier,
		},
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving identities")
	}

	var result []string
	for _, identity := range identities {
		if !identity.Pubkey.Valid {
			// If one of the identities does not have a public key
			// then invitation should be made through an invitation link
			// so let's just not expose the public keys
			return make([]string, 0), nil
		} else {
			result = append(result, identity.Pubkey.String)
		}
	}

	return result, nil
}
