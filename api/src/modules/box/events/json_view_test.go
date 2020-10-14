package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func TestFormatEvent(t *testing.T) {
	t.Run("the deleted content is optional", func(t *testing.T) {
		// prepare the event
		msg := MsgTextContent{
			Encrypted: "coucou",
			PublicKey: "pub-key",
		}
		var json types.JSON
		err := json.Marshal(msg)
		assert.Nil(t, err)
		e := Event{
			Type:        "msg.text",
			JSONContent: json,
			SenderID:    "user-A",
		}

		// we need to set into mem the identity a because the querier is not mocked
		identities := NewIdentityMapper(nil)
		identities.mem["user-A"] = SenderView{DisplayName: "Jin"}

		_, err = e.Format(context.Background(), identities, false)
		assert.Nil(t, err)
	})

	t.Run("the deletor is correctly mapped", func(t *testing.T) {
		// prepare the event
		msg := DeletedContent{}
		msg.Deleted.ByIdentityID = "delelor-A"
		var json types.JSON
		err := json.Marshal(msg)
		assert.Nil(t, err)
		e := Event{
			Type:        "msg.text",
			JSONContent: json,
			SenderID:    "delelor-A",
		}
		// we need to set into mem the identity a because the querier is not mocked
		identities := NewIdentityMapper(nil)
		identities.mem["delelor-A"] = SenderView{DisplayName: "Arno"}

		view, err := e.Format(context.Background(), identities, false)
		assert.Nilf(t, err, "to view")
		assert.Equal(t, view.Sender.DisplayName, "Arno")

		err = e.JSONContent.Marshal(&msg)
		assert.Nilf(t, err, "marshal deleted content")
	})

	t.Run("the kicker is correctly mapped", func(t *testing.T) {
		// prepare the event
		msg := MemberKickContent{}
		msg.KickerID = "kicker-A"
		var json types.JSON
		err := json.Marshal(msg)
		assert.Nil(t, err)
		e := Event{
			Type:        "member.kick",
			JSONContent: json,
			SenderID:    "kicker-A",
		}
		// we need to set into mem the identity a because the querier is not mocked
		identities := NewIdentityMapper(nil)
		identities.mem["kicker-A"] = SenderView{DisplayName: "Floufy"}

		view, err := e.Format(context.Background(), identities, false)
		assert.Nilf(t, err, "to view")
		assert.Equal(t, view.Sender.DisplayName, "Floufy")

		err = e.JSONContent.Marshal(&msg)
		assert.Nilf(t, err, "marshal member kick content")
	})
}
