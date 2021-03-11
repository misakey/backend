package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// SenderView is how an event sender (or box creator)
// is represented in JSON reponses
type SenderView struct {
	ID              string      `json:"id"`
	DisplayName     string      `json:"display_name"`
	AvatarURL       null.String `json:"avatar_url"`
	IdentifierValue string      `json:"identifier_value"`
	IdentifierKind  string      `json:"identifier_kind"`

	accountID       null.String
	identityPubkeys identity.IdentityPublicKeys
}

func (sender SenderView) copyOpaque() SenderView {
	sender.IdentifierValue = ""
	sender.IdentifierKind = ""
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

// ToJSON ...
func (v *View) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// Format event into its JSON view.
func (e Event) Format(ctx context.Context, identityMapper *IdentityMapper, transparent bool) (View, error) {
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
	view.Sender, err = identityMapper.Get(ctx, e.SenderID, transparent)
	if err != nil {
		return view, merr.From(err).Desc("getting sender")
	}

	// For deleted messages
	// we put the deletor identifier in the content
	if e.Type == etype.Msgtext || e.Type == etype.Msgfile {
		var content DeletedContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return view, merr.From(err).Descf("unmarshaling %s json", e.Type)
		}

		if content.Deleted.ByIdentityID != "" { // only if identityMapper is set then bind deletor
			deletor, err := identityMapper.Get(ctx, content.Deleted.ByIdentityID, transparent)
			if err != nil {
				return view, merr.From(err).Desc("getting delelor")
			}
			content.Deleted.ByIdentity = &deletor
			content.Deleted.ByIdentityID = ""
			if err := view.Content.Marshal(content); err != nil {
				return view, merr.From(err).Descf("marshalling %s content", e.Type)
			}
		}
	}

	if e.Type == etype.Memberkick {
		var content MemberKickContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return view, merr.From(err).Descf("unmarshaling %s json", e.Type)
		}

		if content.KickerID != "" {
			kicker, err := identityMapper.Get(ctx, content.KickerID, transparent)
			if err != nil {
				return view, merr.From(err).Desc("getting kicker")
			}
			content.Kicker = &kicker
			content.KickerID = ""
		}

		if err := view.Content.Marshal(content); err != nil {
			return view, merr.From(err).Desc("marshalling event content")
		}
	}

	if view.Content.String() == "{}" {
		view.Content = nil
	}
	return view, nil
}
