package consent

// Context bears internal data about current user consent request
type Context struct {
	Subject        string   `json:"subject"`
	Challenge      string   `json:"challenge"`
	Skip           bool     `json:"skip"`
	ACR            string   `json:"acr"`
	RequestedScope []string `json:"requested_scope"`
	// Context        map[string]string `json:"context"`
}

// Acceptance contains data about the user consent approval
type Acceptance struct {
	GrantScope  []string `json:"grant_scope"`
	Remember    bool     `json:"remember"`
	RememberFor int      `json:"remember_for"`
	Session     struct {
		IDTokenClaims struct {
			// extra self-contained claims on ID Token
			Scope string `json:"sco"`
			Email string `json:"email"`
		} `json:"id_token"`
		AccessTokenClaims struct {
			// extra instropection claims on Access Token
			ACR string `json:"acr"`
		} `json:"access_token"`
	} `json:"session"`
}

// Redirect information for the user's agent
type Redirect struct {
	To string `json:"redirect_to"`
}
