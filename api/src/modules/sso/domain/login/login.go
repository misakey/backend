package login

// Context bears internal data about current user authentication request
type Context struct {
	Challenge      string   `json:"challenge"`
	Skip           bool     `json:"skip"`
	Subject        string   `json:"subject"`
	RequestedScope []string `json:"requested_scope"`
	Client         struct { // concerned relying party
		ID string `json:"client_id"`
	} `json:"client"`
	OIDCContext struct { // OIDC context of the current request
		ACRValues []string `json:"acr_values"`
		LoginHint string   `json:"login_hint"`
	} `json:"oidc_context"`
}

// FlowInfo bears data about current user authentication status that we can share externally
// It is in a way the Context structure with less data
type FlowInfo struct {
	ClientID       string   `json:"client_id"`
	RequestedScope []string `json:"scope"`
	ACRValues      []string `json:"acr_values"`
	LoginHint      string   `json:"login_hint"`
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}

// Acceptance contains data about the user authentication approval
type Acceptance struct {
	Subject     string `json:"subject"`
	ACR         string `json:"acr"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
}

// // LogoutRequest contains the id of the user
// type LogoutRequest struct {
// 	UserID string `json:"user_id" validate:"required"`
// }
