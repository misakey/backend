package events

// Content are metadata stored as JSON in an event.
// Depending of the type of event, the content answers a
// defined format. This file contains logic about event's content:
// formmating, validation...

import (
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// EmptyContent ...
type EmptyContent struct{}

// Unmarshal ...
func (c *EmptyContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

// Validate ...
func (c EmptyContent) Validate() error {
	return nil
}

type anyContent interface {
	// Unmarshal should allow the JSON encoding of the content
	// into a types.JSON variable
	Unmarshal(types.JSON) error

	// Validate should confirm the content contains all expected data
	// in an appropriate format
	Validate() error
}

var contentTypeGetters = map[string]func() anyContent{
	"access.add":        func() anyContent { return &accessAddContent{} },
	"create":            func() anyContent { return &CreationContent{} },
	"msg.text":          func() anyContent { return &MsgTextContent{} },
	"msg.file":          func() anyContent { return &MsgFileContent{} },
	"msg.edit":          func() anyContent { return &MsgEditContent{} },
	"state.access_mode": func() anyContent { return &AccessModeContent{} },
}

func bindAndValidateContent(e *Event) error {
	contentTypeGet, ok := contentTypeGetters[e.Type]
	if !ok {
		// trick to avoid problems when coming
		// to unmarshal an "empty" json type
		if e.JSONContent.String() == "" {
			_ = e.JSONContent.Marshal(&EmptyContent{})
		}
		return nil
	}

	c := contentTypeGet()
	if err := e.JSONContent.Unmarshal(c); err != nil {
		return merr.BadRequest().Descf("unmarshalling %s: %v", e.Type, err)
	}
	return c.Validate()
}
