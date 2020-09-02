package events

// Content are metadata stored as JSON in an event.
// Depending of the type of event, the content answers a
// defined format. This file contains logic about event's content:
// formmating, validation...

import (
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var contentTypeGetters = map[string]func() anyContent{
	"create":          func() anyContent { return &CreationContent{} },
	"state.lifecycle": func() anyContent { return &StateLifecycleContent{} },
	"msg.text":        func() anyContent { return &MsgTextContent{} },
	"msg.file":        func() anyContent { return &MsgFileContent{} },
	"msg.delete":      func() anyContent { return &MsgDeleteContent{} },
	"msg.edit":        func() anyContent { return &MsgEditContent{} },
	"join":            func() anyContent { return &JoinContent{} },
}

type anyContent interface {
	// Unmarshal should allow the JSON encoding of the content
	// into a types.JSON variable
	Unmarshal(types.JSON) error

	// Validate should confirm the content contains all expected data
	// in an appropriate format
	Validate() error
}

func bindAndValidateContent(e *Event) error {
	contentTypeGet, ok := contentTypeGetters[e.Type]
	if !ok {
		return merror.Internal().Describef("unknown content type %s", e.Type)
	}

	e.Content = contentTypeGet()
	if err := e.Content.Unmarshal(e.JSONContent); err != nil {
		return err
	}
	return e.Content.Validate()
}
