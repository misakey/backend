package quota

import (
	"context"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

// Quotum model
type Quotum struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	IdentityID string    `json:"identity_id"`
	Value      int64     `json:"value"`
	Origin     string    `json:"origin"`
}

// ToDomain ...
func ToDomain(dbQuotum sqlboiler.StorageQuotum) Quotum {
	return Quotum{
		ID:         dbQuotum.ID,
		CreatedAt:  dbQuotum.CreatedAt,
		IdentityID: dbQuotum.IdentityID,
		Value:      dbQuotum.Value,
		Origin:     dbQuotum.Origin,
	}
}

// ToSQLBoiler ...
func (q Quotum) ToSQLBoiler() *sqlboiler.StorageQuotum {
	return &sqlboiler.StorageQuotum{
		ID:         q.ID,
		CreatedAt:  q.CreatedAt,
		IdentityID: q.IdentityID,
		Value:      q.Value,
		Origin:     q.Origin,
	}
}

// Create quotum generating the id
func Create(ctx context.Context, exec boil.ContextExecutor, quotum *Quotum) error {
	var err error
	quotum.ID, err = uuid.NewString()
	if err != nil {
		return merr.From(err).Desc("generating uuid")
	}
	return quotum.ToSQLBoiler().Insert(ctx, exec, boil.Infer())
}

// List quota for a given identityID
func List(ctx context.Context, exec boil.ContextExecutor, id string) ([]Quotum, error) {
	dbQuota, err := sqlboiler.StorageQuota(sqlboiler.StorageQuotumWhere.IdentityID.EQ(id)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	if len(dbQuota) == 0 {
		return []Quotum{}, nil
	}

	quota := make([]Quotum, len(dbQuota))
	for idx, quotum := range dbQuota {
		quota[idx] = ToDomain(*quotum)
	}
	return quota, nil
}
