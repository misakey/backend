package authn

import (
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
)

// Step in a multi-factor authentication
// Today, an authentication can only be one step.
type Step struct {
	ID              int
	IdentityID      string     `json:"identity_id"`
	MethodName      MethodRef  `json:"method_name"`
	RawJSONMetadata types.JSON `json:"metadata"`
	CreatedAt       time.Time
	Complete        bool
	CompleteAt      null.Time
}
