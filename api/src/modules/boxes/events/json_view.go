package events

import (
	"time"

	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

// View represent an event as it is represented in JSON responses
type View struct {
	// TODO factorize with other type definitions
	Type      string     `json:"type"`
	Content   types.JSON `json:"content"`
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"server_event_created_at"`
	Sender    Sender     `json:"sender"`
}

// ToView transforms an event into its JSON view.
func ToView(e *Event, senderIdentity *domain.Identity) View {
	view := View{
		Type:      e.Type,
		Content:   e.Content,
		ID:        e.ID,
		CreatedAt: e.CreatedAt,
		Sender:    Sender(senderIdentity.DisplayName),
	}

	return view
}
