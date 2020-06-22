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
	"create":          func() anyContent { return &creationContent{} },
	"state.lifecycle": func() anyContent { return &stateLifecycleContent{} },
	"msg.text":        func() anyContent { return &msgTextContent{} },
	"msg.file":        func() anyContent { return &msgFileContent{} },
}

type anyContent interface {
	// Unmarshal should allow the JSON encoding of the content
	// into a types.JSON variable
	Unmarshal(types.JSON) error

	// Validate should confirm the content contains all expected data
	// in an appropriate format
	Validate() error
}

func validateContent(e Event) error {
	contentTypeGet, ok := contentTypeGetters[e.Type]
	if !ok {
		return merror.Internal().Describef("unknown content type %s", e.Type)
	}

	content := contentTypeGet()
	if err := content.Unmarshal(e.Content); err != nil {
		return err
	}
	return content.Validate()
}
