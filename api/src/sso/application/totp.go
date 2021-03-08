package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mtotp"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// EnrollmentView ...
type EnrollmentView struct {
	ID       string `json:"id"`
	B64Image string `json:"base64_image"`
}

// BeginTOTPEnrollmentQuery ...
type BeginTOTPEnrollmentQuery struct {
	identityID string
}

// BindAndValidate ...
func (cmd *BeginTOTPEnrollmentQuery) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// BeginTOTPEnrollment returns options to register webauthn credentials
func (sso *SSOService) BeginTOTPEnrollment(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BeginTOTPEnrollmentQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	curIdentity, err := identity.Get(ctx, sso.ssoDB, query.identityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identityID")
	}

	// TOTP options
	opts := totp.GenerateOpts{
		Issuer:      sso.AuthenticationService.AppName,
		AccountName: curIdentity.IdentifierValue,
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return nil, merr.From(err).Desc("generating totp key")
	}

	// encode image in base64
	var buf bytes.Buffer
	img, err := key.Image(180, 180)
	if err != nil {
		return nil, merr.From(err).Desc("generating qr code")
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, merr.From(err).Desc("encoding image")
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	id, err := uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("generating id")
	}

	view := EnrollmentView{
		ID:       id,
		B64Image: encoded,
	}

	// the cache is only valid for 10 minutes
	// then the user must ask for a new secret
	if _, err := sso.redConn.Set(fmt.Sprintf("totp:%s", id), key.Secret(), 10*time.Minute).Result(); err != nil {
		return nil, merr.From(err).Desc("storing secret")
	}

	return view, nil
}

// RecoveryCodesView ...
type RecoveryCodesView struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// FinishTOTPEnrollmentQuery ...
type FinishTOTPEnrollmentQuery struct {
	Code       string `json:"code"`
	ID         string `json:"id"`
	identityID string
}

// BindAndValidate ...
func (cmd *FinishTOTPEnrollmentQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// FinishTOTPEnrollment returns options to register webauthn credentials
func (sso *SSOService) FinishTOTPEnrollment(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*FinishTOTPEnrollmentQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	secret, err := sso.redConn.Get(fmt.Sprintf("totp:%s", query.ID)).Result()
	if err != nil && err == redis.Nil {
		return nil, merr.Forbidden().Desc("expired or non existing secret")
	}
	if err != nil {
		return nil, merr.From(err).Desc("getting secret")
	}

	if !totp.Validate(query.Code, secret) {
		return nil, merr.Forbidden()
	}

	recoveryCodes, err := mtotp.GenerateRecoveryCodes()
	if err != nil {
		return nil, merr.From(err).Desc("generating recovery codes")
	}

	toStore := sqlboiler.TotpSecret{
		IdentityID: query.identityID,
		Secret:     secret,
		Backup:     recoveryCodes,
	}
	if err := toStore.Insert(ctx, sso.ssoDB, boil.Infer()); err != nil {
		return nil, merr.From(err).Desc("storing secret")
	}
	if _, err := sso.redConn.Del(fmt.Sprintf("totp:%s", query.ID)).Result(); err != nil {
		logger.FromCtx(ctx).Err(err).Msgf("deleting %s redis key", fmt.Sprintf("totp:%s", query.ID))
	}

	return RecoveryCodesView{
		RecoveryCodes: toStore.Backup,
	}, nil
}

// RegenerateRecoveryCodesQuery ...
type RegenerateRecoveryCodesQuery struct {
	identityID string
}

// BindAndValidate ...
func (cmd *RegenerateRecoveryCodesQuery) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// RegenerateRecoveryCodes returns options to register webauthn credentials
func (sso *SSOService) RegenerateRecoveryCodes(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*RegenerateRecoveryCodesQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	secret, err := sqlboiler.TotpSecrets(sqlboiler.TotpSecretWhere.IdentityID.EQ(query.identityID)).One(ctx, sso.ssoDB)
	if err != nil {
		return nil, err
	}
	secret.Backup, err = mtotp.GenerateRecoveryCodes()
	if err != nil {
		return nil, err
	}
	if _, err := secret.Update(ctx, sso.ssoDB, boil.Whitelist(sqlboiler.TotpSecretColumns.Backup)); err != nil {
		return []string{}, err
	}

	return RecoveryCodesView{
		RecoveryCodes: secret.Backup,
	}, nil
}

// DeleteSecretQuery ...
type DeleteSecretQuery struct {
	identityID string
}

// BindAndValidate ...
func (cmd *DeleteSecretQuery) BindAndValidate(eCtx echo.Context) error {
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// DeleteSecret for a given identity
func (sso *SSOService) DeleteSecret(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*DeleteSecretQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	// check that the operation can be performed
	curIdentity, err := identity.Get(ctx, sso.ssoDB, query.identityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identityID")
	}

	if curIdentity.MFAMethod == "totp" {
		return nil, merr.Conflict().Desc("mfa_method is totp")
	}

	// delete the secret
	rowsAff, err := sqlboiler.TotpSecrets(sqlboiler.TotpSecretWhere.IdentityID.EQ(query.identityID)).DeleteAll(ctx, sso.ssoDB)
	if err != nil {
		return nil, err
	}
	if rowsAff == 0 {
		return nil, merr.NotFound()
	}

	return nil, nil
}
