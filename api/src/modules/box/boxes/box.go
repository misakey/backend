package boxes

import (
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
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
