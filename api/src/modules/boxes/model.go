package boxes

import (
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/events"
)

type boxState struct {
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}
type Box struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"server_created_at"`
	Creator   events.SenderView `json:"creator"`
	boxState
}
