package authn

import (
	"context"
	"database/sql"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type accountMetadata struct {
	Password      argon2.HashedPassword         `json:"prehashed_password"`
	BackupData    string                        `json:"backup_data"`
	SecretStorage crypto.SecretStorageSetupData `json:"secret_storage"`
}

// Validate ...
func (am accountMetadata) Validate() error {
	if err := v.ValidateStruct(&am,
		v.Field(&am.Password),
	); err != nil {
		return merr.From(err).Desc("validating account metadata")
	}

	// TODO IN TRAIL OF FLOWERS stop accepting backup data at all
	if am.BackupData == "" {
		if err := am.SecretStorage.Validate(); err != nil {
			return merr.From(err).Desc("validating account secret storage")
		}
	}

	return nil
}

// assertAccountCreation
func (as *Service) assertAccountCreation(
	ctx context.Context, tr *sql.Tx, redConn *redis.Client,
	challenge string, curIdentity *identity.Identity, step Step,
) error {
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.ACR.LessThan(oidc.ACR1) {
		return merr.Forbidden()
	}

	if acc.Subject != challenge {
		return merr.Forbidden().Ori(merr.OriHeaders).
			Desc("authorization header must correspond to the login_challenge").
			Add("Authorization", merr.DVConflict).Add("login_challenge", merr.DVConflict)
	}

	if acc.IdentityID != curIdentity.ID {
		return merr.Forbidden().Desc("wrong identity id")
	}

	// transform metadata into account metadata structure
	accountMetadata, err := toMetadata(step.RawJSONMetadata)
	if err != nil {
		return merr.BadRequest().
			Desc(err.Error()).Add("metadata", merr.DVMalformed)
	}

	if err := accountMetadata.Validate(); err != nil {
		return merr.From(err).Desc("validating account metadata")
	}

	if curIdentity.AccountID.Valid {
		return merr.Forbidden().Desc("identity has already an account")
	}

	account := identity.Account{}

	if accountMetadata.BackupData != "" {
		// should not happen in production,
		// only useful for testing the migration of old accounts to the new system
		// TODO IN TRAIL OF FLOWERS remove this
		account.BackupData = accountMetadata.BackupData
		account.BackupVersion = 1
	}

	// hash the password before storing it
	account.Password, err = accountMetadata.Password.Hash()
	if err != nil {
		return merr.From(err).Desc("could not hash the password")
	}
	if err := identity.CreateAccount(ctx, tr, &account); err != nil {
		return err
	}

	if accountMetadata.BackupData == "" {
		err = crypto.ResetAccountSecretStorage(ctx, tr, account.ID, &accountMetadata.SecretStorage)
		if err != nil {
			return merr.From(err).Desc("setting up secret storage")
		}
		err = curIdentity.SetIdentityKeys(
			accountMetadata.SecretStorage.IdentityPublicKey,
			accountMetadata.SecretStorage.IdentityNonIdentifiedPublicKey,
		)
		if err != nil {
			return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
		}
	}

	// update the identity's account id column
	curIdentity.AccountID = null.StringFrom(account.ID)
	if err := identity.Update(ctx, tr, curIdentity); err != nil {
		return err
	}

	// create identity notification about account creation
	if err := identity.NotificationCreate(
		ctx, tr, redConn,
		curIdentity.ID, "user.create_account", null.JSONFromPtr(nil),
	); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("notifying identity %s", curIdentity.ID)
	}
	return nil
}

func toMetadata(msg types.JSON) (ret accountMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, merr.From(err).Desc("password metadata")
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
