package mwebauthn

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

func CredentialsNumber(ctx context.Context, exec boil.ContextExecutor, identityID string) (int64, error) {
	mods := []qm.QueryMod{
		sqlboiler.WebauthnCredentialWhere.IdentityID.EQ(identityID),
	}

	return sqlboiler.WebauthnCredentials(mods...).Count(ctx, exec)
}

func CredentialsExist(ctx context.Context, exec boil.ContextExecutor, identityID string) bool {
	number, err := CredentialsNumber(ctx, exec, identityID)
	if err != nil {
		return false
	}

	return number > 0
}
