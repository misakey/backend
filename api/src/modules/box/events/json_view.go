package events

import (
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

// SenderView is how an event sender (or box creator)
// is represented in JSON reponses
type SenderView struct {
	DisplayName string      `json:"display_name"`
	AvatarURL   null.String `json:"avatar_url"`
	Identifier  struct {
		Value string `json:"value"`
		Kind  string `json:"kind"`
	} `json:"identifier"`
}

func NewSenderView(identity domain.Identity) SenderView {
	result := SenderView{
		DisplayName: identity.DisplayName,
		AvatarURL:   identity.AvatarURL,
	}

	result.Identifier.Kind = string(identity.Identifier.Kind)
	result.Identifier.Value = identity.Identifier.Value

	return result
}

// View represent an event as it is represented in JSON responses
type View struct {
	Type      string     `json:"type"`
	Content   types.JSON `json:"content"`
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"server_event_created_at"`
	Sender    SenderView `json:"sender"`
}

// ToView transforms an event into its JSON view.
func ToView(e Event, senderIdentity domain.Identity) View {
	view := View{
		Type:      e.Type,
		Content:   e.Content,
		ID:        e.ID,
		CreatedAt: e.CreatedAt,
		Sender:    NewSenderView(senderIdentity),
	}

	return view
}
