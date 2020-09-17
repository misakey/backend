package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

func TestFormatEvent(t *testing.T) {
	t.Run("the deleted content is optional", func(t *testing.T) {
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
		_, err = FormatEvent(e, map[string]domain.Identity{})
		assert.Nil(t, err)
	})
	t.Run("the sender mapping of deleted content is optional", func(t *testing.T) {
		msg := DeletedContent{}
		msg.Deleted.ByIdentityID = "identity-A"

		var json types.JSON
		err := json.Marshal(msg)
		assert.Nil(t, err)
		e := Event{
			Type:        "msg.text",
			JSONContent: json,
			SenderID:    "user-A",
		}
		_, err = FormatEvent(e, map[string]domain.Identity{})
		assert.Nil(t, err)
	})
	t.Run("the deletor is correctly mapped", func(t *testing.T) {
		msg := DeletedContent{}
		msg.Deleted.ByIdentityID = "identity-A"

		var json types.JSON
		err := json.Marshal(msg)
		assert.Nil(t, err)
		e := Event{
			Type:        "msg.text",
			JSONContent: json,
			SenderID:    "identity-A",
		}
		view, err := FormatEvent(e, map[string]domain.Identity{
			"identity-A": domain.Identity{DisplayName: "Arno"},
		})
		assert.Nilf(t, err, "to view")
		assert.Equal(t, view.Sender.DisplayName, "Arno")

		err = e.JSONContent.Marshal(&msg)
		assert.Nilf(t, err, "marshal deleted content")
	})
}
