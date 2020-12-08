package authflow

// BuildResetURL ...
func (afs Service) BuildResetURL(authURL string) string {
	// by default (no authURL found), return the home page URL
	if authURL == "" {
		return afs.homePageURL.String()
	}
	return authURL
}
