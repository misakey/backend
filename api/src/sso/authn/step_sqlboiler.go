package authn

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

func createStep(ctx context.Context, exec boil.ContextExecutor, step *Step) error {
	// convert domain to sql model
	sqlAuthnStep := sqlboiler.AuthenticationStep{
		IdentityID: step.IdentityID,
		MethodName: string(step.MethodName),
		Metadata:   step.RawJSONMetadata,
	}
	if !step.CreatedAt.IsZero() {
		sqlAuthnStep.CreatedAt = step.CreatedAt
	}
	fmt.Println("the step, ", step)
	if err := sqlAuthnStep.Insert(ctx, exec, boil.Infer()); err != nil {
		return err
	}
	fmt.Println("inserted ! ", sqlAuthnStep.ID)
	// copy data potentially created in SQL layer
	step.ID = sqlAuthnStep.ID
	step.CreatedAt = sqlAuthnStep.CreatedAt
	return nil
}

func completeAtStep(ctx context.Context, exec boil.ContextExecutor, id int, completeTime time.Time) error {
	data := sqlboiler.M{sqlboiler.AuthenticationStepColumns.CompleteAt: null.TimeFrom(completeTime)}
	rowsAff, err := sqlboiler.AuthenticationSteps(sqlboiler.AuthenticationStepWhere.ID.EQ(id)).UpdateAll(ctx, exec, data)
	if rowsAff == 0 {
		return merror.NotFound().Describe("no rows affected in persistent layer")
	}
	return err
}

func getLastStep(
	ctx context.Context, exec boil.ContextExecutor,
	identityID string, methodName oidc.MethodRef,
) (Step, error) {

	authnStep := Step{}

	mods := []qm.QueryMod{
		sqlboiler.AuthenticationStepWhere.IdentityID.EQ(identityID),
		sqlboiler.AuthenticationStepWhere.MethodName.EQ(string(methodName)),
		qm.OrderBy(sqlboiler.AuthenticationStepColumns.CreatedAt + " DESC"),
	}

	sqlAuthnStep, err := sqlboiler.AuthenticationSteps(mods...).One(ctx, exec)
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
	authnStep.MethodName = oidc.MethodRef(sqlAuthnStep.MethodName)
	authnStep.RawJSONMetadata = sqlAuthnStep.Metadata
	authnStep.CreatedAt = sqlAuthnStep.CreatedAt
	authnStep.CompleteAt = sqlAuthnStep.CompleteAt
	authnStep.Complete = sqlAuthnStep.CompleteAt.Valid
	return authnStep, nil
}

func deleteIncompleteSteps(ctx context.Context, exec boil.ContextExecutor, identityID string) error {
	mods := []qm.QueryMod{
		sqlboiler.AuthenticationStepWhere.IdentityID.EQ(identityID),
		sqlboiler.AuthenticationStepWhere.CompleteAt.IsNull(),
	}
	// ignore no rows affected since not incomplete step deleted means
	// no incomplete steps anymore in storage: the method did its job
	_, err := sqlboiler.AuthenticationSteps(mods...).DeleteAll(ctx, exec)
	return err
}
