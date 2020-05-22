package domain

import (
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
)

type IdentityProof struct {
	ID         string
	IdentityID string
	MethodName IdentityProofingMethod `json:"method_name"`
	Metadata   json.RawMessage        `json:"metadata"`

	InitiatedAt time.Time

	Asserted   bool
	AssertedAt null.Time
}

type IdentityProofingMethod string

const (
	EmailCodeMethod IdentityProofingMethod = "email_confirmation_code"
)

type IdentityProofFilters struct {
	IdentityID *string
	MethodName *IdentityProofingMethod
	Asserted   null.Bool
	LastFirst  null.Bool
}
