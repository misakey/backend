package domain

type Identifier struct {
	ID    string         `json:"id"`
	Value string         `json:"value"`
	Kind  IdentifierKind `json:"kind"`
}

type IdentifierKind string

const (
	EmailIdentifier IdentifierKind = "email"
)
