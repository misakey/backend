package domain

type Identity struct {
	ID            string `json:"id"`
	AccountID     string `json:"account_id"`
	IdentifierID  string `json:"identifer_id"`
	IsAuthable    bool   `json:"is_authable"`
	DisplayName   string `json:"display_name"`
	Notifications string `json:"notifications"`
	AvatarURL     string `json:"avatar_url"`
	Confirmed     bool   `json:"confirmed"`
}
