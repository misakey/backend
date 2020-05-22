package domain

type Identifier struct {
	ID    string
	Value string
	Kind  IdentifierKind
}

type IdentifierKind string

const (
	EmailIdentifier IdentifierKind = "email"
)
