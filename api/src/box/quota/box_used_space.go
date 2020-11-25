package quota

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

type UsedSpace struct {
	BoxID string `json:"box_id"`
	Value int64  `json:"value"`
	ID    string `json:"-"`
}

func usedSpaceToDomain(dbBoxUsedSpace sqlboiler.BoxUsedSpace) UsedSpace {
	return UsedSpace{
		BoxID: dbBoxUsedSpace.BoxID,
		Value: dbBoxUsedSpace.Value,
	}
}

func ListBoxUsedSpaces(ctx context.Context, exec boil.ContextExecutor, boxIds []string) ([]UsedSpace, error) {
	dbBoxUsedSpace, err := sqlboiler.BoxUsedSpaces(sqlboiler.BoxUsedSpaceWhere.BoxID.IN(boxIds)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	if len(dbBoxUsedSpace) == 0 {
		return []UsedSpace{}, nil
	}

	boxUsedSpace := make([]UsedSpace, len(dbBoxUsedSpace))
	for idx, usedSpace := range dbBoxUsedSpace {
		boxUsedSpace[idx] = usedSpaceToDomain(*usedSpace)
	}
	return boxUsedSpace, nil
}

func DeleteBoxUsedSpace(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	// rowAff is ignored because this is called on delete box and it should not fail if no used space
	// was existing for the box
	_, err := sqlboiler.BoxUsedSpaces(sqlboiler.BoxUsedSpaceWhere.BoxID.EQ(boxID)).DeleteAll(ctx, exec)
	return err
}

func UpdateBoxUsedSpace(ctx context.Context, exec boil.ContextExecutor, BoxID string, incrementValue int64, decrementValue int64) error {
	// retrieve the current boxUsedSpace
	currentBoxUsedSpace, err := sqlboiler.BoxUsedSpaces(sqlboiler.BoxUsedSpaceWhere.BoxID.EQ(BoxID)).One(ctx, exec)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// generate a new uuid as a new box used space ID
		id, err := uuid.NewString()
		if err != nil {
			return merror.Transform(err).Describe("generating new box used space id")
		}
		value := incrementValue - decrementValue
		if value < 0 {
			value = 0
		}
		boxUsedSpace := sqlboiler.BoxUsedSpace{
			ID:    id,
			BoxID: BoxID,
			Value: value,
		}
		return boxUsedSpace.Insert(ctx, exec, boil.Infer())
	}

	newValue := currentBoxUsedSpace.Value + incrementValue - decrementValue
	if newValue < 0 {
		newValue = 0
	}
	currentBoxUsedSpace.Value = newValue

	_, err = currentBoxUsedSpace.Update(ctx, exec, boil.Infer())
	return err
}
