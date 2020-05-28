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
	Confirmed     bool        `json:"confirmed"`
}

type IdentityFilters struct {
	IdentifierID null.String
	IsAuthable   null.Bool
}
