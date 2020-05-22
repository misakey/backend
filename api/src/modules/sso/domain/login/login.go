package login

// Acceptance contains data about authentication information approval (from the system)
type Acceptance struct {
	Subject     string `json:"subject"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
}

// Context bears internal data about current user authentication request
type Context struct {
	Skip           bool     `json:"skip"`
	Subject        string   `json:"subject"`
	RequestedScope []string `json:"requested_scope"`
	Client         struct { // concerned relying party
		ID string `json:"client_id"`
	} `json:"client"`
	OIDCContext struct { // OIDC context of the current request
		LoginHint string `json:"login_hint"`
	} `json:"oidc_context"`
}

// // PublicInfo bears data about current user authentication status that we can share externally
// // It is in a way the LoginContext structure with less data
// type PublicInfo struct {
// 	ClientID       string   `json:"client_id"`
// 	RequestedScope []string `json:"scope"`
// 	ACRValues      []string `json:"acr_values"`
// 	LoginHint      string   `json:"login_hint"`
// }

// Redirect information for the client
type Redirect struct {
	To string `json:"redirect_to"`
}

// // LogoutRequest contains the id of the user
// type LogoutRequest struct {
// 	UserID string `json:"user_id" validate:"required"`
// }
