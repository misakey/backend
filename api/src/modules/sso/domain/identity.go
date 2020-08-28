package domain

import "github.com/volatiletech/null"

type Identity struct {
	ID            string      `json:"id"`
	AccountID     null.String `json:"account_id"`
	IdentifierID  string      `json:"identifier_id"`
	IsAuthable    bool        `json:"is_authable"`
	DisplayName   string      `json:"display_name"`
	Notifications string      `json:"notifications"`
	AvatarURL     null.String `json:"avatar_url"`
	Color         null.String       `json:"color"`

	// Identifier is always returned within the identity entity as a nested JSON object
	Identifier Identifier `json:"identifier"`
}

type IdentityFilters struct {
	IdentifierID null.String
	IsAuthable   null.Bool
	IDs          []string
	AccountID    null.String
}
