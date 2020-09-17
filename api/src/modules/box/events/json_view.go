package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
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
	Type       string      `json:"type"`
	Content    *types.JSON `json:"content"`
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"server_event_created_at"`
	ReferrerID null.String `json:"referrer_id"`
	Sender     SenderView  `json:"sender"`
}

func (v *View) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

func ToView(e Event, sender domain.Identity) View {
	return View{
		Type:       e.Type,
		Content:    &e.JSONContent,
		ID:         e.ID,
		ReferrerID: e.ReferrerID,
		CreatedAt:  e.CreatedAt,
		Sender:     NewSenderView(sender),
	}
}

// FormatEvent transforms an event into its JSON view.
func FormatEvent(e Event, identityMap map[string]domain.Identity) (View, error) {
	view := ToView(e, identityMap[e.SenderID])

	// For deleted messages
	// we put the deletor identifier in the content
	if e.Type == "msg.text" || e.Type == "msg.file" {
		var content DeletedContent
		err := json.Unmarshal(e.JSONContent, &content)
		if err != nil {
			return view, merror.Transform(err).Describef("unmarshaling %s json", e.Type)
		}

		if content.Deleted.ByIdentityID != "" {
			deletor := NewSenderView(identityMap[content.Deleted.ByIdentityID])
			content.Deleted.ByIdentity = &deletor
			content.Deleted.ByIdentityID = ""
			err := view.Content.Marshal(content)
			if err != nil {
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

		if content.KickedMemberID != "" {
			kicked := NewSenderView(identityMap[content.KickedMemberID])
			content.KickedMember = &kicked
		}
		content.KickedMemberID = ""
		if err := view.Content.Marshal(content); err != nil {
			return view, merror.Transform(err).Describe("marshalling event content")
		}
	}

	if view.Content.String() == "{}" {
		view.Content = nil
	}
	return view, nil
}

func MapSenderIdentities(ctx context.Context, events []Event, identityRepo entrypoints.IdentityIntraprocessInterface) (map[string]domain.Identity, error) {
	// getting senders IDs without duplicates
	// we build both a set (map to bool) and a list
	// because we need a list in "domain.IdentityFilters"
	var senderIDs []string
	idMap := make(map[string]bool)
	for _, event := range events {
		IDsInEvent, err := getIdentityIDs(event)
		if err != nil {
			return nil, merror.Transform(err).Describe("getting identity IDs in event")
		}

		for _, ID := range IDsInEvent {
			_, alreadyPresent := idMap[ID]
			if !alreadyPresent {
				senderIDs = append(senderIDs, ID)
				idMap[ID] = true
			}
		}
	}

	identities, err := identityRepo.List(ctx, domain.IdentityFilters{IDs: senderIDs})
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving Identities from IDs")
	}

	var sendersMap = make(map[string]domain.Identity, len(identities))
	for _, identity := range identities {
		sendersMap[identity.ID] = *identity
	}

	return sendersMap, nil
}

func getIdentityIDs(event Event) ([]string, error) {
	ids := []string{event.SenderID}

	if event.Type == "msg.text" || event.Type == "msg.file" {
		var content DeletedContent
		err := json.Unmarshal(event.JSONContent, &content)
		if err != nil {
			return ids, merror.Transform(err).Describef("unmarshaling %s json", event.Type)
		}

		if content.Deleted.ByIdentityID != "" {
			ids = append(ids, content.Deleted.ByIdentityID)
		}
	}
	if event.Type == "member.kick" {
		var content MemberKickContent
		err := json.Unmarshal(event.JSONContent, &content)
		if err != nil {
			return ids, merror.Transform(err).Describef("unmarshaling %s json", event.Type)
		}

		if content.KickedMemberID != "" {
			ids = append(ids, content.KickedMemberID)
		}
	}

	return ids, nil
}
