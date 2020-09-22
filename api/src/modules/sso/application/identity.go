package application

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type IdentityQuery struct {
	IdentityID string
}

func (query IdentityQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.IdentityID, v.Required, is.UUIDv4.Error("identity id must be an uuid v4")),
	)
}

type IdentityView struct {
	domain.Identity
}

func (sso SSOService) GetIdentity(ctx context.Context, query IdentityQuery) (IdentityView, error) {
	view := IdentityView{}
	var err error

	// verify identity access
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden()
	}

	if acc.IdentityID != query.IdentityID {
		return view, merror.Forbidden()
	}

	// set the view on identity retrieval
	view.Identity, err = sso.identityService.Get(ctx, query.IdentityID)
	if err != nil {
		return view, err
	}
	return view, err
}

// PartialUpdateIdentityCmd
type PartialUpdateIdentityCmd struct {
	IdentityID    string
	DisplayName   string      `json:"display_name"`
	Notifications string      `json:"notifications"`
	Color         null.String `json:"color"`
}

// Validate the IdentityAuthableCmd
func (cmd PartialUpdateIdentityCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Notifications, v.In("minimal", "moderate", "frequent")),
		v.Field(&cmd.DisplayName, v.Length(3, 254)),
		v.Field(&cmd.Color, v.Length(7, 7)),
	); err != nil {
		return err
	}
	return nil
}

// PartialUpdateIdentity to change its display name or avatar
func (sso *SSOService) PartialUpdateIdentity(ctx context.Context, cmd PartialUpdateIdentityCmd) error {
	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.IdentityID {
		return merror.Forbidden()
	}

	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	if cmd.DisplayName != "" {
		identity.DisplayName = cmd.DisplayName
	}

	if cmd.Notifications != "" {
		identity.Notifications = cmd.Notifications
	}

	if cmd.Color.Valid {
		identity.Color = cmd.Color
	}

	return sso.identityService.Update(ctx, &identity)
}

// UploadAvatarCmd
type UploadAvatarCmd struct {
	IdentityID string
	Data       io.Reader
	Extension  string
}

// Validate the UploadAvatarCmd
func (cmd UploadAvatarCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Extension, v.Required),
	); err != nil {
		return err
	}
	return nil
}

func (sso *SSOService) UploadAvatar(ctx context.Context, cmd UploadAvatarCmd) error {
	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.IdentityID {
		return merror.Forbidden()
	}

	// get avatar's corresponding user
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	// first remove existing avatar if it does exist
	if !identity.AvatarURL.IsZero() {
		avatarToDel := &domain.AvatarFile{
			Filename: filepath.Base(identity.AvatarURL.String),
		}
		if err := sso.identityService.DeleteAvatar(ctx, avatarToDel); err != nil {
			return err
		}
	}

	avatar := domain.AvatarFile{}
	// generate an UUID to use as a filename
	filename, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	avatar.Filename = filename.String() + avatar.Extension
	avatar.Extension = cmd.Extension
	avatar.Data = cmd.Data

	// upload the avatar to storage
	url, err := sso.identityService.UploadAvatar(ctx, &avatar)
	if err != nil {
		return err
	}

	// then save into user account the new avatar uri
	identity.AvatarURL = null.StringFrom(url)
	return sso.identityService.Update(ctx, &identity)
}

// DeleteAvatarCmd
type DeleteAvatarCmd struct {
	IdentityID string
}

// Validate the DeleteAvatarCmd
func (cmd DeleteAvatarCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
	)
}

// DeleteAvatar for a given identity
func (sso *SSOService) DeleteAvatar(ctx context.Context, cmd DeleteAvatarCmd) error {
	avatar := domain.AvatarFile{}

	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.IdentityID {
		return merror.Forbidden()
	}

	// get identity
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	// check that the identity has an avatar to delet
	if identity.AvatarURL.IsZero() {
		return merror.Conflict().Describe("avatar is not set").Detail("identity_id", merror.DVConflict)
	}

	// delete avatar
	avatar.Filename = filepath.Base(identity.AvatarURL.String)
	err = sso.identityService.DeleteAvatar(ctx, &avatar)
	if err != nil {
		return err
	}

	// update identity with empty avatar url field
	identity.AvatarURL = null.NewString("", false)

	return sso.identityService.Update(ctx, &identity)
}

// AttachCouponCmd
type AttachCouponCmd struct {
	IdentityID string
	Value      string
}

// Validate the AttachCouponCmd
func (cmd AttachCouponCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Value, v.Required),
	)
}

// AttachCoupon to a given identity
func (sso *SSOService) AttachCoupon(ctx context.Context, cmd AttachCouponCmd) error {
	acc := ajwt.GetAccesses(ctx)

	// verify requested user id and authenticated user id are the same.
	if acc == nil || acc.IdentityID != cmd.IdentityID {
		return merror.Forbidden()
	}

	// get identity
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return err
	}

	// Check the coupon hasn't already been applied
	if identity.Level >= 20 {
		return merror.Conflict().Describe("coupon already applied")
	}

	if strings.EqualFold(cmd.Value, "ProductHuntRocks") {
		cmd.Value = "ProductHuntRocks"
	} else if strings.EqualFold(cmd.Value, "EarlyBird") || strings.EqualFold(cmd.Value, "EarlyBirds") {
		cmd.Value = "EarlyBird"
	} else {
		return merror.BadRequest().Detail("value", merror.DVInvalid).Detail("invalid_value", cmd.Value).Describe("invalid coupon")
	}

	// 3. Update the identity
	identity.Level = 20
	if err := sso.identityService.Update(ctx, &identity); err != nil {
		return merror.Transform(err).Describe("updating identity")
	}

	// 2. Create the used_coupon
	usedCoupon := domain.UsedCoupon{
		IdentityID: cmd.IdentityID,
		Value:      cmd.Value,
	}

	if err := sso.usedCouponService.CreateUsedCoupon(ctx, usedCoupon); err != nil {
		return merror.Transform(err).Describe("creating coupon")
	}

	return nil
}
