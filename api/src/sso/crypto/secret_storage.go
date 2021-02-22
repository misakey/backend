package crypto

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

type AccountRootKey struct {
	KeyHash      string `json:"key_hash"`
	EncryptedKey string `json:"encrypted_key"`
}

func (k *AccountRootKey) Validate() error {
	if err := v.ValidateStruct(k,
		v.Field(&k.KeyHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&k.EncryptedKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return err
	}

	return nil
}

type VaultKey struct {
	KeyHash      string `json:"key_hash"`
	EncryptedKey string `json:"encrypted_key"`
}

func (k *VaultKey) Validate() error {
	if err := v.ValidateStruct(k,
		v.Field(&k.KeyHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&k.EncryptedKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return err
	}

	return nil
}

type SecretStorageAsymKey struct {
	PublicKey          string `json:"public_key,omitempty"`
	EncryptedSecretKey string `json:"encrypted_secret_key"`
	AccountRootKeyHash string `json:"account_root_key_hash,omitempty"`
}

// BindAndValidate implements request.Request.BindAndValidate
func (asymKey *SecretStorageAsymKey) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(asymKey); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	if err := v.ValidateStruct(asymKey,
		v.Field(&asymKey.PublicKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&asymKey.AccountRootKeyHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&asymKey.EncryptedSecretKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return err
	}

	return nil
}

type SecretStorageBoxKeyShare struct {
	ID                       string    `json:"id"`
	InvitationShareHash      string    `json:"invitation_share_hash,omitempty"`
	EncryptedInvitationShare string    `json:"encrypted_invitation_share"`
	AccountRootKeyHash       string    `json:"account_root_key_hash,omitempty"`
	BoxID                    string    `json:"box_id,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (share *SecretStorageBoxKeyShare) Validate() error {
	return v.ValidateStruct(share,
		v.Field(&share.ID, v.Empty),
		v.Field(&share.InvitationShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&share.EncryptedInvitationShare, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&share.AccountRootKeyHash, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&share.BoxID, v.Required, is.UUIDv4),
		v.Field(&share.CreatedAt, v.Empty),
		v.Field(&share.UpdatedAt, v.Empty),
	)
}

// BindAndValidate implements request.Request.BindAndValidate
func (share *SecretStorageBoxKeyShare) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(share); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	share.BoxID = eCtx.Param("box-id")

	return share.Validate()
}

type Secrets struct {
	AccountRootKey AccountRootKey                      `json:"account_root_key"`
	VaultKey       VaultKey                            `json:"vault_key"`
	AsymKeys       map[string]SecretStorageAsymKey     `json:"asym_keys"`
	BoxKeyShares   map[string]SecretStorageBoxKeyShare `json:"box_key_shares"`
}

var (
	ErrNoRootKey = errors.New("no root key")
)

func GetCurrentAccountRootKey(ctx context.Context, exec boil.ContextExecutor, accountID string) (*sqlboiler.SecretStorageAccountRootKey, error) {
	dbRootKey, err := sqlboiler.SecretStorageAccountRootKeys(
		sqlboiler.SecretStorageAccountRootKeyWhere.AccountID.EQ(accountID),
		qm.OrderBy(sqlboiler.SecretStorageAccountRootKeyColumns.CreatedAt+" DESC"),
	).One(ctx, exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbRootKey, ErrNoRootKey
		}
		return dbRootKey, err
	}

	return dbRootKey, nil

}

func GetVaultKeyByRootKeyHash(ctx context.Context, exec boil.ContextExecutor, rootKeyHash string) (*sqlboiler.SecretStorageVaultKey, error) {
	dbVaultKey, err := sqlboiler.SecretStorageVaultKeys(
		sqlboiler.SecretStorageVaultKeyWhere.AccountRootKeyHash.EQ(rootKeyHash),
	).One(ctx, exec)
	if err != nil {
		return dbVaultKey, err
	}

	return dbVaultKey, nil

}

func GetAccountSecrets(ctx context.Context, exec boil.ContextExecutor, accountID string) (Secrets, error) {
	result := Secrets{}

	dbRootKey, err := GetCurrentAccountRootKey(ctx, exec, accountID)
	if err != nil {
		return result, err
	}
	result.AccountRootKey.KeyHash = dbRootKey.KeyHash
	result.AccountRootKey.EncryptedKey = dbRootKey.EncryptedKey

	rootKeyHash := dbRootKey.KeyHash

	dbVaultKey, err := GetVaultKeyByRootKeyHash(ctx, exec, rootKeyHash)
	if err != nil {
		return result, err
	}
	result.VaultKey.KeyHash = dbVaultKey.KeyHash
	result.VaultKey.EncryptedKey = dbVaultKey.EncryptedKey

	dbAsymKeys, err := sqlboiler.SecretStorageAsymKeys(
		sqlboiler.SecretStorageAsymKeyWhere.AccountRootKeyHash.EQ(rootKeyHash),
	).All(ctx, exec)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}
	if err != sql.ErrNoRows {
		result.AsymKeys = make(map[string]SecretStorageAsymKey, 1)
		for _, dbAsymKey := range dbAsymKeys {
			result.AsymKeys[dbAsymKey.PublicKey] = SecretStorageAsymKey{
				EncryptedSecretKey: dbAsymKey.EncryptedSecretKey,
			}
		}
	}

	dbBoxKeyShares, err := sqlboiler.SecretStorageBoxKeyShares(
		sqlboiler.SecretStorageBoxKeyShareWhere.AccountRootKeyHash.EQ(rootKeyHash),
	).All(ctx, exec)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}
	// XXX if there are no rows, are we sure it's okay
	// to have non-initialized map?
	if err != sql.ErrNoRows {
		result.BoxKeyShares = make(map[string]SecretStorageBoxKeyShare, 1)
		for _, dbBoxKeyShare := range dbBoxKeyShares {
			result.BoxKeyShares[dbBoxKeyShare.BoxID] = SecretStorageBoxKeyShare{
				ID:                       dbBoxKeyShare.ID,
				EncryptedInvitationShare: dbBoxKeyShare.EncryptedInvitationShare,
				InvitationShareHash:      dbBoxKeyShare.InvitationShareHash,
				CreatedAt:                dbBoxKeyShare.CreatedAt,
				UpdatedAt:                dbBoxKeyShare.UpdatedAt,
			}
		}
	}

	return result, nil
}

func UpdateRootKey(ctx context.Context, exec boil.ContextExecutor, accountID string, encryptedKey string) error {
	dbRootKey, err := GetCurrentAccountRootKey(ctx, exec, accountID)
	if err != nil {
		return err
	}

	dbRootKey.EncryptedKey = encryptedKey

	nbRowsAffected, err := dbRootKey.Update(ctx, exec, boil.Infer())
	if nbRowsAffected != 1 {
		return errors.New(`updating account encrypted root key: unexpected nb of affected rows (${nbRowsAffected})`)
	}
	if err != nil {
		return err
	}

	return nil
}

// SecretStorageSetupData ...
type SecretStorageSetupData struct {
	AccountRootKey                 AccountRootKey                      `json:"account_root_key"`
	VaultKey                       VaultKey                            `json:"vault_key"`
	AsymKeys                       map[string]SecretStorageAsymKey     `json:"asym_keys"`
	BoxKeyShares                   map[string]SecretStorageBoxKeyShare `json:"box_key_shares"`
	IdentityPublicKey              string                              `json:"identity_public_key"`
	IdentityNonIdentifiedPublicKey string                              `json:"identity_non_identified_public_key"`
}

// BindAndValidate implements request.Request.BindAndValidate
func (query *SecretStorageSetupData) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	if err := query.Validate(); err != nil {
		return err
	}

	return nil
}

func (query *SecretStorageSetupData) Validate() error {
	if err := query.AccountRootKey.Validate(); err != nil {
		return merr.From(err).Desc("validating root key")
	}

	if err := query.VaultKey.Validate(); err != nil {
		return merr.From(err).Desc("validating vault key")
	}

	for publicKey, asymKey := range query.AsymKeys {
		if err := v.Validate(publicKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)); err != nil {
			return err
		}

		if err := v.ValidateStruct(&asymKey,
			v.Field(&asymKey.EncryptedSecretKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
			v.Field(&asymKey.PublicKey, v.Empty),
		); err != nil {
			return err
		}
	}

	for boxID, keyShare := range query.BoxKeyShares {
		keyShare.BoxID = boxID

		if err := keyShare.Validate(); err != nil {
			return err
		}
	}

	// identity keys are required
	// *unless* we are migrating an account which already has some
	err := v.ValidateStruct(query,
		v.Field(&query.IdentityPublicKey, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&query.IdentityNonIdentifiedPublicKey, v.Match(format.UnpaddedURLSafeBase64)),
	)
	if err != nil {
		return err
	}

	return nil
}

// ResetAccountSecretStorage creates a brand new secret storage for the given account.
// Use it for account creation, password reset
// and migration from the old "secret backup" system to the secret storage system
func ResetAccountSecretStorage(ctx context.Context, tr *sql.Tx, accountID string, data *SecretStorageSetupData) error {
	dbRootKey := sqlboiler.SecretStorageAccountRootKey{
		AccountID:    accountID,
		KeyHash:      data.AccountRootKey.KeyHash,
		EncryptedKey: data.AccountRootKey.EncryptedKey,
	}

	if err := dbRootKey.Insert(ctx, tr, boil.Infer()); err != nil {
		return err
	}

	dbVaultKey := sqlboiler.SecretStorageVaultKey{
		AccountRootKeyHash: data.AccountRootKey.KeyHash,
		KeyHash:            data.VaultKey.KeyHash,
		EncryptedKey:       data.VaultKey.EncryptedKey,
	}

	if err := dbVaultKey.Insert(ctx, tr, boil.Infer()); err != nil {
		return err
	}

	for publicKey, asymKey := range data.AsymKeys {
		dbAsymKey := sqlboiler.SecretStorageAsymKey{
			PublicKey:          publicKey,
			EncryptedSecretKey: asymKey.EncryptedSecretKey,
			AccountRootKeyHash: data.AccountRootKey.KeyHash,
		}
		var err error
		dbAsymKey.ID, err = uuid.NewString()
		if err != nil {
			return merr.From(err).Desc("generating uuid")
		}
		if err := dbAsymKey.Insert(ctx, tr, boil.Infer()); err != nil {
			return err
		}
	}

	for boxID, boxKeyShare := range data.BoxKeyShares {
		dbBoxKeyShare := sqlboiler.SecretStorageBoxKeyShare{
			InvitationShareHash:      boxKeyShare.InvitationShareHash,
			EncryptedInvitationShare: boxKeyShare.EncryptedInvitationShare,
			BoxID:                    boxID,
			AccountRootKeyHash:       data.AccountRootKey.KeyHash,
		}
		var err error
		dbBoxKeyShare.ID, err = uuid.NewString()
		if err != nil {
			return merr.From(err).Desc("generating uuid")
		}
		if err := dbBoxKeyShare.Insert(ctx, tr, boil.Infer()); err != nil {
			return merr.From(err).Desc("inserting box key share")
		}
	}

	return nil
}

func CreateSecretStorageAsymKey(ctx context.Context, exec boil.ContextExecutor, accountID string, asymKey SecretStorageAsymKey) (*sqlboiler.SecretStorageAsymKey, error) {
	dbRootKey, err := GetCurrentAccountRootKey(ctx, exec, accountID)
	if err != nil {
		return nil, err
	}

	if asymKey.AccountRootKeyHash != dbRootKey.KeyHash {
		// either the client is using an old root key for this account,
		// or it is trying to mess up with someone else's secret storage
		// (one of our wors threats regarding end-to-end encryption)
		return nil, merr.Forbidden().Ori("account_root_key_hash").
			Desc("account root key hash does not match key hash of current account root key")
	}

	dbAsymKey := sqlboiler.SecretStorageAsymKey{
		PublicKey:          asymKey.PublicKey,
		EncryptedSecretKey: asymKey.EncryptedSecretKey,
		// We could also take `asymKey.AccountRootKeyHash` given that we just checked that they are equal
		// but let's take any opportunity we can to limit the amount of user-controlled data,
		// especially for one of the most sensitive field of the data
		AccountRootKeyHash: dbRootKey.KeyHash,
	}
	dbAsymKey.ID, err = uuid.NewString()
	if err != nil {
		return nil, merr.From(err).Desc("generating uuid")
	}
	if err := dbAsymKey.Insert(ctx, exec, boil.Infer()); err != nil {
		sqlError, ok := merr.Cause(err).(*pq.Error)
		if ok && sqlError.Constraint == "one_per_pubkey_per_root_key" {
			// If the frontend attempted to overwrite a public key we do nothing but we don't return an error
			return &dbAsymKey, nil
		}
		return nil, err
	}

	return &dbAsymKey, nil
}

func CreateSecretStorageBoxKeyShare(ctx context.Context, tr *sql.Tx, accountID string, share SecretStorageBoxKeyShare) (*sqlboiler.SecretStorageBoxKeyShare, error) {
	dbRootKey, err := GetCurrentAccountRootKey(ctx, tr, accountID)
	if err != nil {
		return nil, err
	}

	if share.AccountRootKeyHash != dbRootKey.KeyHash {
		// either the client is using an old root key for this account,
		// or it is trying to mess up with someone else's secret storage
		// (one of our worst threats regarding end-to-end encryption)
		return nil, merr.Forbidden().Ori("account_root_key_hash").
			Desc("account root key hash does not match key hash of current account root key")
	}

	var result *sqlboiler.SecretStorageBoxKeyShare

	shareToOverwrite, err := sqlboiler.SecretStorageBoxKeyShares(
		sqlboiler.SecretStorageBoxKeyShareWhere.BoxID.EQ(share.BoxID),
		sqlboiler.SecretStorageBoxKeyShareWhere.AccountRootKeyHash.EQ(dbRootKey.KeyHash),
	).One(ctx, tr)
	if err != nil && err != sql.ErrNoRows {
		return nil, merr.From(err).Desc("getting share to overwrite")
	}

	if err == sql.ErrNoRows {
		result = &sqlboiler.SecretStorageBoxKeyShare{
			InvitationShareHash:      share.InvitationShareHash,
			EncryptedInvitationShare: share.EncryptedInvitationShare,
			BoxID:                    share.BoxID,
			AccountRootKeyHash:       dbRootKey.KeyHash,
		}

		result.ID, err = uuid.NewString()
		if err != nil {
			return nil, merr.From(err).Desc("generating uuid")
		}

		if err := result.Insert(ctx, tr, boil.Infer()); err != nil {
			return nil, merr.From(err).Desc("inserting new row")
		}
	} else {
		result = shareToOverwrite
		result.InvitationShareHash = share.InvitationShareHash
		result.EncryptedInvitationShare = share.EncryptedInvitationShare

		nbRowsAffected, err := result.Update(ctx, tr, boil.Infer())
		if err != nil {
			return nil, merr.From(err).Desc("updating existing row")
		}
		if nbRowsAffected != 1 {
			return nil, merr.From(err).Descf(`updating existing row: expected 1 affected row, got %d`, nbRowsAffected)
		}
	}

	return result, nil
}

type RootKeyShare struct {
	AccountID      string `json:"account_id"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

func CreateRootKeyShare(ctx context.Context, redConn *redis.Client, rootKeyShare RootKeyShare, expirationTime time.Duration) error {
	key := "rootkeyshare:" + rootKeyShare.OtherShareHash
	value, err := json.Marshal(rootKeyShare)
	if err != nil {
		return merr.Internal().Desc("encoding root key share").Desc(err.Error())
	}
	if _, err := redConn.Set(key, value, expirationTime).Result(); err != nil {
		return err
	}
	return nil
}

func GetRootKeyShare(ctx context.Context, redConn *redis.Client, otherShareHash string) (*RootKeyShare, error) {
	key := "rootkeyshare:" + otherShareHash
	value, err := redConn.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, merr.NotFound()
		}
		return nil, err
	}

	share := RootKeyShare{}
	if err := json.Unmarshal([]byte(value), &share); err != nil {
		return nil, merr.From(err).Desc("unmarshaling redis value")
	}

	return &share, nil
}

func DeleteAsymKeys(ctx context.Context, tr *sql.Tx, accountID string, pubkeys []string) error {
	accountRootKey, err := GetCurrentAccountRootKey(ctx, tr, accountID)
	if err != nil {
		return merr.From(err).Desc("retrieving current account root key")
	}

	nbRowsAffected, err := sqlboiler.SecretStorageAsymKeys(
		sqlboiler.SecretStorageAsymKeyWhere.PublicKey.IN(pubkeys),
		sqlboiler.SecretStorageAsymKeyWhere.AccountRootKeyHash.EQ(accountRootKey.KeyHash),
	).DeleteAll(ctx, tr)

	if err != nil {
		return err
	}
	if nbRowsAffected < int64(len(pubkeys)) {
		// disabled until we can detect rows that were already deleted
		// (see https://gitlab.misakey.dev/misakey/backend/-/issues/289)
		// return merr.NotFound().Descf("only %d rows deleted (expected %d)", nbRowsAffected, len(pubkeys))
	} else if nbRowsAffected > int64(len(pubkeys)) {
		return merr.Conflict().Descf("%d rows deleted (expected %d)", nbRowsAffected, len(pubkeys))
	}

	return nil
}

func DeleteBoxKeyShares(ctx context.Context, tr *sql.Tx, accountID string, boxIDs []string) error {
	accountRootKey, err := GetCurrentAccountRootKey(ctx, tr, accountID)
	if err != nil {
		return merr.From(err).Desc("retrieving current account root key")
	}

	nbRowsAffected, err := sqlboiler.SecretStorageBoxKeyShares(
		sqlboiler.SecretStorageBoxKeyShareWhere.BoxID.IN(boxIDs),
		sqlboiler.SecretStorageBoxKeyShareWhere.AccountRootKeyHash.EQ(accountRootKey.KeyHash),
	).DeleteAll(ctx, tr)

	if err != nil {
		return err
	}
	if nbRowsAffected < int64(len(boxIDs)) {
		// disabled until we can detect rows that were already deleted
		// (see https://gitlab.misakey.dev/misakey/backend/-/issues/289)
		// return merr.NotFound().Descf("only %d rows deleted (expected %d)", nbRowsAffected, len(boxIDs))
	} else if nbRowsAffected > int64(len(boxIDs)) {
		return merr.Conflict().Descf("%d rows deleted (expected %d)", nbRowsAffected, len(boxIDs))
	}

	return nil
}
