package authn

import (
	"context"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type accountMetadata struct {
	Password   argon2.HashedPassword `json:"prehashed_password"`
	BackupData string                `json:"backup_data"`
}

func (am accountMetadata) Validate() error {
	if err := v.ValidateStruct(&am,
		v.Field(&am.Password),
		v.Field(&am.BackupData, v.Required),
	); err != nil {
		return merror.Transform(err).Describe("validating account metadata")
	}
	return nil
}

// assertAccountCreation
func (as *Service) assertAccountCreation(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	challenge string, curIdentity *identity.Identity, step Step,
) error {
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.ACR.LessThan(oidc.ACR1) {
		return merror.Forbidden()
	}

	if acc.Subject != challenge {
		return merror.Forbidden().From(merror.OriHeaders).
			Describe("authorization header must correspond to the login_challenge").
			Detail("Authorization", merror.DVConflict).Detail("login_challenge", merror.DVConflict)
	}

	if acc.IdentityID != curIdentity.ID {
		return merror.Forbidden().Describe("wrong identity id")
	}

	// transform metadata into account metadata structure
	accountMetadata, err := toMetadata(step.RawJSONMetadata)
	if err != nil {
		return merror.BadRequest().
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}

	if err := accountMetadata.Validate(); err != nil {
		return merror.Transform(err).Describe("validating account metadata")
	}

	if curIdentity.AccountID.Valid {
		return merror.Forbidden().Describe("identity has already an account")
	}

	// prepare the account to be created
	account := identity.Account{
		BackupData:    accountMetadata.BackupData,
		BackupVersion: 1,
	}

	// hash the password before storing it
	account.Password, err = accountMetadata.Password.Hash()
	if err != nil {
		return merror.Transform(err).Describe("could not hash the password")
	}
	if err := identity.CreateAccount(ctx, exec, &account); err != nil {
		return err
	}

	// update the identity's account id column
	curIdentity.AccountID = null.StringFrom(account.ID)
	if err := identity.Update(ctx, exec, curIdentity); err != nil {
		return err
	}

	// set initial quotas
	_, err = as.quotaService.CreateBase(ctx, curIdentity.ID)
	if err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("setting base quota for %s", curIdentity.ID)
	}

	// create identity notification about account creation
	if err := identity.NotificationCreate(
		ctx, exec, redConn,
		curIdentity.ID, "user.create_account", null.JSONFromPtr(nil),
	); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("notifying identity %s", curIdentity.ID)
	}
	return nil
}

func toMetadata(msg types.JSON) (ret accountMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, merror.Transform(err).Describe("password metadata")
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
