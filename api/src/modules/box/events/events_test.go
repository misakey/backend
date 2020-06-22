package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWithAnyContent(t *testing.T) {
	content := msgContent{
		Encrypted: "bm90X2VtcHR5X3lvdV9zZWU=",
	}
	e, err := NewWithAnyContent("msg.text", &content, "3389043f-bf0a-456c-a8a2-f068ede21ce9", "2289043f-bf0a-456c-a8a2-f068ede21ce9")
	assert.Equalf(t, "msg.text", e.Type, "event type")
	assert.Equalf(t, "2289043f-bf0a-456c-a8a2-f068ede21ce9", e.SenderID, "sender id")
	assert.Equalf(t, "3389043f-bf0a-456c-a8a2-f068ede21ce9", e.BoxID, "box id")
	assert.Equalf(t, "{\"encrypted\":\"bm90X2VtcHR5X3lvdV9zZWU=\"}", e.Content.String(), "content")
	assert.NotEmptyf(t, e.ID, "empty id")
	assert.WithinDurationf(t, time.Now(), e.CreatedAt, time.Second, "created_at")
	assert.Nilf(t, err, "error not nil")
}
