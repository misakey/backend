package mtotp

import (
	"context"
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mrand"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

func SecretExist(ctx context.Context, exec boil.ContextExecutor, identityID string) bool {
	mods := []qm.QueryMod{
		sqlboiler.TotpSecretWhere.IdentityID.EQ(identityID),
	}

	exists, _ := sqlboiler.TotpSecrets(mods...).Exists(ctx, exec)

	return exists
}

func CheckAndDeleteRecoveryCode(ctx context.Context, exec boil.ContextExecutor, secretID int, code string) error {
	// here we want atomic queries to avoid problems with concurrent accesses
	mods := []qm.QueryMod{
		sqlboiler.TotpSecretWhere.ID.EQ(secretID),
		qm.Where("?=ANY(backup)", code),
	}

	exist, err := sqlboiler.TotpSecrets(mods...).Exists(ctx, exec)
	if err != nil {
		return err
	}
	if !exist {
		return merr.Forbidden().Ori(merr.OriBody).Add("metadata", merr.DVInvalid)
	}

	// remove the code in a single query to keep on beeing atomic
	if _, err := queries.Raw(fmt.Sprintf("UPDATE %s SET backup = array_remove(backup, $1) WHERE id=$2", sqlboiler.TableNames.TotpSecret), code, secretID).ExecContext(ctx, exec); err != nil {
		return err
	}

	return nil
}

func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)

	for i := 0; i < 10; i++ {
		leftPart, err := mrand.String(5)
		if err != nil {
			return codes, err
		}
		rightPart, err := mrand.String(5)
		if err != nil {
			return codes, err
		}
		codes[i] = fmt.Sprintf("%s-%s", leftPart, rightPart)
	}

	return codes, nil
}
