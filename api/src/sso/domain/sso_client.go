package domain

import "github.com/volatiletech/null/v8"

// SSOClient ...
type SSOClient struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	LogoURL null.String `json:"logo_url"`
}
