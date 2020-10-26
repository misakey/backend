package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// SenderView is how an event sender (or box creator)
// is represented in JSON reponses
type SenderView struct {
	ID           string      `json:"id"`
	DisplayName  string      `json:"display_name"`
	AvatarURL    null.String `json:"avatar_url"`
	IdentifierID string      `json:"identifier_id"`
	Identifier   struct {
		Value string `json:"value"`
		Kind  string `json:"kind"`
	} `json:"identifier"`
}

func (sender SenderView) copyOpaque() SenderView {
	sender.Identifier.Value = ""
	sender.Identifier.Kind = ""
	return sender
}

// View represent an event as it is represented in JSON responses
type View struct {
	Type       string      `json:"type"`
	Content    *types.JSON `json:"content"`
	BoxID      string      `json:"box_id"`
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"server_event_created_at"`
	ReferrerID null.String `json:"referrer_id"`
	Sender     SenderView  `json:"sender"`
}

func (v *View) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// FormatEvent transforms an event into its JSON view.
func (e Event) Format(ctx context.Context, identities *IdentityMapper, transparent bool) (View, error) {
	view := View{
		Type:       e.Type,
		Content:    &e.JSONContent,
		ID:         e.ID,
		BoxID:      e.BoxID,
		ReferrerID: e.ReferrerID,
		CreatedAt:  e.CreatedAt,
	}
	// map the sender
	var err error
	view.Sender, err = identities.Get(ctx, e.SenderID, transparent)
	if err != nil {
		return view, merror.Transform(err).Describe("getting sender")
	}

	// For deleted messages
	// we put the deletor identifier in the content
	if e.Type == "msg.text" || e.Type == "msg.file" {
		var content DeletedContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return view, merror.Transform(err).Describef("unmarshaling %s json", e.Type)
		}

		if content.Deleted.ByIdentityID != "" {
			deletor, err := identities.Get(ctx, content.Deleted.ByIdentityID, transparent)
			if err != nil {
				return view, merror.Transform(err).Describe("getting delelor")
			}
			content.Deleted.ByIdentity = &deletor
			content.Deleted.ByIdentityID = ""
			if err := view.Content.Marshal(content); err != nil {
				return view, merror.Transform(err).Describef("marshalling %s content", e.Type)
			}
		}
	}

	if e.Type == "member.kick" {
		var content MemberKickContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return view, merror.Transform(err).Describef("unmarshaling %s json", e.Type)
		}

		if content.KickerID != "" {
			kicker, err := identities.Get(ctx, content.KickerID, transparent)
			if err != nil {
				return view, merror.Transform(err).Describe("getting kicker")
			}
			content.Kicker = &kicker
		}
		content.KickerID = ""
		if err := view.Content.Marshal(content); err != nil {
			return view, merror.Transform(err).Describe("marshalling event content")
		}
	}

	if view.Content.String() == "{}" {
		view.Content = nil
	}
	return view, nil
}
