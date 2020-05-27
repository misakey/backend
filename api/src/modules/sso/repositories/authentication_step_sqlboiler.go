package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type AuthenticationStepSQLBoiler struct {
	db *sql.DB
}

func NewAuthenticationStepSQLBoiler(db *sql.DB) *AuthenticationStepSQLBoiler {
	return &AuthenticationStepSQLBoiler{
		db: db,
	}
}

func (repo *AuthenticationStepSQLBoiler) Create(ctx context.Context, authnStep *authentication.Step) error {
	// convert domain to sql model
	sqlAuthnStep := sqlboiler.AuthenticationStep{
		IdentityID: authnStep.IdentityID,
		MethodName: string(authnStep.MethodName),
		Metadata:   types.JSON(authnStep.Metadata),
	}
	if !authnStep.CreatedAt.IsZero() {
		sqlAuthnStep.CreatedAt = authnStep.CreatedAt
	}
	if err := sqlAuthnStep.Insert(ctx, repo.db, boil.Infer()); err != nil {
		return err
	}
	// copy data potentially created in SQL layer
	authnStep.ID = sqlAuthnStep.ID
	authnStep.CreatedAt = sqlAuthnStep.CreatedAt
	return nil
}

func (repo *AuthenticationStepSQLBoiler) CompleteAt(ctx context.Context, id int, completeTime time.Time) error {
	data := sqlboiler.M{sqlboiler.AuthenticationStepColumns.CompleteAt: null.TimeFrom(completeTime)}
	rowsAff, err := sqlboiler.AuthenticationSteps(sqlboiler.AuthenticationStepWhere.ID.EQ(id)).UpdateAll(ctx, repo.db, data)
	if rowsAff == 0 {
		return merror.NotFound().Describe("no rows affected in persistent layer")
	}
	return err
}

func (repo *AuthenticationStepSQLBoiler) Last(
	ctx context.Context,
	identityID string,
	methodName authentication.Method,
) (authentication.Step, error) {

	authnStep := authentication.Step{}

	mods := []qm.QueryMod{
		sqlboiler.AuthenticationStepWhere.IdentityID.EQ(identityID),
		sqlboiler.AuthenticationStepWhere.MethodName.EQ(string(methodName)),
		qm.OrderBy(sqlboiler.AuthenticationStepColumns.CreatedAt),
	}

	sqlAuthnStep, err := sqlboiler.AuthenticationSteps(mods...).One(ctx, repo.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return authnStep, merror.NotFound().Describe(err.Error()).
				Detail("identity_id", merror.DVNotFound).
				Detail("method_name", merror.DVNotFound)
		}
		return authnStep, err
	}

	// build domain model based on sql data
	authnStep.ID = sqlAuthnStep.ID
	authnStep.IdentityID = sqlAuthnStep.IdentityID
	authnStep.MethodName = authentication.Method(sqlAuthnStep.MethodName)
	authnStep.Metadata = sqlAuthnStep.Metadata
	authnStep.CreatedAt = sqlAuthnStep.CreatedAt
	authnStep.CompleteAt = sqlAuthnStep.CompleteAt
	authnStep.Complete = sqlAuthnStep.CompleteAt.Valid
	return authnStep, nil
}
