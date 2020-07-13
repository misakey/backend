package authn

// Session is bound to login session id in hydra
// it has the same ID and is expired automatically
// the same moment as hydra's session
// RememberFor is expressed in seconds
type Session struct {
	ID          string
	ACR         ClassRef
	RememberFor int
}
