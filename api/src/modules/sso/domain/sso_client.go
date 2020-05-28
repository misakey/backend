package domain

import "github.com/volatiletech/null"

type SSOClient struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	LogoURL null.String `json:"logo_url"`
}
