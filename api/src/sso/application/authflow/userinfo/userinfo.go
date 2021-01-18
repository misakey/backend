package userinfo

// UserInfo basically bears the ID Token info
type UserInfo struct {
	ACR   string   `json:"acr"`
	AID   string   `json:"aid"`
	AMR   []string `json:"amr"`
	Email string   `json:"email"`
	MID   string   `json:"mid"`
	Sco   string   `json:"sco"`
	SID   string   `json:"sid"`
	Sub   string   `json:"sub"`
}
