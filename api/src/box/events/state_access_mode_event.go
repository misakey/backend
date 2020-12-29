package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// enum for access mode
const (
	PublicMode  = "public"
	LimitedMode = "limited"
)

type AccessModeContent struct {
	Value string `json:"value"`
}

// Unmarshal ...
func (c *AccessModeContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

// Validate ...
func (c AccessModeContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Value, v.Required, v.In(PublicMode, LimitedMode)),
	)
}

func doStateAccessMode(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, _ *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check accesses
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}
	return nil, e.persist(ctx, exec)
}
