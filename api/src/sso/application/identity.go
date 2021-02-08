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

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mtotp"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mwebauthn"
)

// IdentityQuery ...
type IdentityQuery struct {
	identityID string
}

// BindAndValidate ...
func (query *IdentityQuery) BindAndValidate(eCtx echo.Context) error {
	query.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(query,
		v.Field(&query.identityID, v.Required, is.UUIDv4),
	); err != nil {
		return merr.From(err).Desc("validating identity query")
	}
	return nil
}

// IdentityView ...
type IdentityView struct {
	identity.Identity
	HasAccount    bool `json:"has_account"`
	HasTOTPSecret bool `json:"has_totp_secret"`
}

// GetIdentity ...
func (sso *SSOService) GetIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityQuery)
	view := IdentityView{}
	var err error

	// verify identity access
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return view, merr.Forbidden()
	}

	if acc.IdentityID != query.identityID {
		return view, merr.Forbidden()
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

	// add information about MFA configuration
	view.HasTOTPSecret = mtotp.SecretExist(ctx, sso.sqlDB, query.identityID)

	return view, err
}

// IdentityPubkeyByIdentifierQuery ...
type IdentityPubkeyByIdentifierQuery struct {
	IdentifierValue string `query:"identifier_value"`
}

// BindAndValidate ...
func (query *IdentityPubkeyByIdentifierQuery) BindAndValidate(eCtx echo.Context) error {
	err := eCtx.Bind(query)
	if err != nil {
		return merr.BadRequest().Desc(err.Error())
	}
	return v.ValidateStruct(query,
		v.Field(&query.IdentifierValue, v.Required),
	)
}

// GetIdentityPubkeyByIdentifier returns a list of pubkeys corresponding to the received identifier
func (sso *SSOService) GetIdentityPubkeyByIdentifier(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*IdentityPubkeyByIdentifierQuery)

	// must only be authenticated
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	identity, err := identity.GetByIdentifierValue(ctx, sso.sqlDB, query.IdentifierValue)
	// return an empty list of not found to respect list route semantic
	if merr.IsANotFound(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, merr.From(err).Desc("retrieving identities")
	}

	if identity.Pubkey.IsZero() {
		return []string{}, nil
	}
	return []string{identity.Pubkey.String}, nil
}

// PartialUpdateIdentityCmd ...
type PartialUpdateIdentityCmd struct {
	identityID          string
	DisplayName         string      `json:"display_name"`
	Notifications       string      `json:"notifications"`
	Color               null.String `json:"color"`
	Pubkey              null.String `json:"pubkey"`
	NonIdentifiedPubkey null.String `json:"non_identified_pubkey"`
	MFAMethod           null.String `json:"mfa_method"`
}

// BindAndValidate the PartialUpdateIdentityCmd
func (cmd *PartialUpdateIdentityCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	cmd.identityID = eCtx.Param("id")
	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Notifications, v.In("minimal", "moderate", "frequent")),
		v.Field(&cmd.DisplayName, v.Length(3, 254)),
		v.Field(&cmd.Color, v.Length(7, 7)),
		v.Field(&cmd.MFAMethod, v.In("disabled", "totp", "webauthn")),
	); err != nil {
		return merr.From(err).Desc("validating identity patch")
	}
	return nil
}

// PartialUpdateIdentity to change its display name or avatar
func (sso *SSOService) PartialUpdateIdentity(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*PartialUpdateIdentityCmd)
	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	curIdentity, err := identity.Get(ctx, tr, cmd.identityID)
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

	if cmd.NonIdentifiedPubkey.Valid {
		curIdentity.NonIdentifiedPubkey = cmd.NonIdentifiedPubkey
	}

	if cmd.MFAMethod.Valid {
		curIdentity.MFAMethod = cmd.MFAMethod.String
		// check if mfa is possible
		if curIdentity.MFAMethod == "webauthn" && !mwebauthn.CredentialsExist(ctx, sso.sqlDB, curIdentity.ID) ||
			curIdentity.MFAMethod == "totp" && !mtotp.SecretExist(ctx, sso.sqlDB, curIdentity.ID) {
			return nil, merr.Conflict().Add("mfa_method", merr.DVConflict).Add("reason", "no credential")
		}
	}

	err = identity.Update(ctx, tr, &curIdentity)
	if err != nil {
		return nil, err
	}

	return nil, tr.Commit()
}

// UploadAvatarCmd ...
type UploadAvatarCmd struct {
	identityID string
	Data       io.Reader
	Extension  string
}

// BindAndValidate ...
func (cmd *UploadAvatarCmd) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")

	file, err := eCtx.FormFile("avatar")
	if err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Add("avatar", merr.DVRequired).Desc(err.Error())
	}
	if file.Size >= 3*1024*1024 {
		return merr.BadRequest().Ori(merr.OriBody).Add("size", merr.DVInvalid).Desc("size must be < 3 mo")
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

// UploadAvatar ...
func (sso *SSOService) UploadAvatar(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*UploadAvatarCmd)
	acc := oidc.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.identityID {
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// get avatar's corresponding user
	existingIdentity, err := identity.Get(ctx, tr, cmd.identityID)
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
	filename, err := uuid.NewRandom()
	if err != nil {
		return nil, merr.From(err).Desc("could not generate uuid v4")
	}
	avatar.Filename = filename.String() + avatar.Extension
	avatar.Extension = cmd.Extension
	avatar.Data = cmd.Data

	// upload the avatar to storage
	url, err := sso.identityService.UploadAvatar(ctx, &avatar)
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

// DeleteAvatarCmd ...
type DeleteAvatarCmd struct {
	identityID string
}

// BindAndValidate ...
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
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// get identity
	curIdentity, err := identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	// check that the identity has an avatar to delet
	if curIdentity.AvatarURL.IsZero() {
		err = merr.Conflict().Desc("avatar is not set").Add("identity_id", merr.DVConflict)
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

// AttachCouponCmd ...
type AttachCouponCmd struct {
	identityID string
	Value      string `json:"value"`
}

// BindAndValidate ...
func (cmd *AttachCouponCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
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
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed across entities
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// get identity
	curIdentity, err := identity.Get(ctx, tr, cmd.identityID)
	if err != nil {
		return nil, err
	}

	// Check the coupon hasn't already been applied
	if curIdentity.Level >= 20 {
		err = merr.Conflict().Desc("coupon already applied")
		return nil, err
	}

	// NOTE: there is no valid coupon nowadays
	err = merr.BadRequest().Add("value", merr.DVInvalid).Add("invalid_value", cmd.Value).Desc("invalid coupon")
	return nil, err
}
