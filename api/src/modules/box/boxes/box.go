package boxes

import (
	"context"
	"time"

	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

// Box is a volatile object built based on events linked to its ID
type Box struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_created_at"`
	PublicKey string    `json:"public_key"`
	Title     string    `json:"title"`
	Lifecycle string    `json:"lifecycle"`

	// aggregated data
	EventsCount int               `json:"events_count"`
	Creator     events.SenderView `json:"creator"`
	LastEvent   events.View       `json:"last_event"`
}

func MustBeAtLeastLevel20(
	ctx context.Context,
	exec boil.ContextExecutor,
	identities entrypoints.IdentityIntraprocessInterface,
	identityID string,
) error {
	identity, err := identities.Get(ctx, identityID)
	if err != nil {
		return merror.Transform(err).Describe("Getting identity")
	}
	if identity.Level < 20 {
		return merror.Forbidden().Detail("level", merror.DVInvalid)
	}
	return nil
}
