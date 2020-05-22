package authentication

import (
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
)

// Step in a multi-factor authentication
// Today, an authentication can only be one step.
type Step struct {
	ID          string
	IdentityID  string          `json:"identity_id"`
	MethodName  Method          `json:"method_name"`
	Metadata    json.RawMessage `json:"metadata"`
	InitiatedAt time.Time

	Complete   bool
	CompleteAt null.Time
}

type Method string

const (
	EmailedCodeMethod Method = "emailed_code"
)
