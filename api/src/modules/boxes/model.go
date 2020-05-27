package boxes

import (
	"time"
)

type Sender struct {
	// Not reusing sqlboiler.Identifier
	// because we may not want to include the ID
	// (XXX do we?)
	Identifer struct {
		Kind  string `json:"kind"`
		Value string `json:"value"`
	} `json:"identifier"`
}

type readOnlyEventFields struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_event_created_at"`
	Sender    Sender    `json:"sender"`
}

type commonEventWritableFields struct {
	Type string `json:"type"`
}

type eventBase struct {
	readOnlyEventFields
	commonEventWritableFields
}

type CreationEvent struct {
	eventBase
	Content boxState `json:"content"`
}

type TextMessageEvent struct {
	eventBase
	Content struct {
		Encrypted string `json:"encrypted"`
	} `json:"content"`
}

type boxState struct {
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}
type Box struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_created_at"`
	Creator   Sender    `json:"creator"`
	boxState
}
