package authn

import (
	"context"
	"time"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/oidc"
)

// Session is bound to login session id in hydra
// it has the same ID and is expired automatically
// the same moment as hydra's session
// RememberFor is expressed in seconds
type Session struct {
	ID          string
	ACR         oidc.ClassRef `json:"acr"`
	IdentityID  string        `json:"mid"`
	AccountID   null.String   `json:"aid"`
	RememberFor int
}

func (as *Service) UpsertSession(ctx context.Context, new Session) error {
	// if session musn't be kept, return directly
	if new.RememberFor <= 0 {
		return nil
	}

	existing, err := as.sessions.Get(ctx, new.ID)
	if err != nil {
		// if not found, we ignore the error to create it
		if !merror.HasCode(err, merror.NotFoundCode) {
			return err
		}
		// if found but the new sec level is inferior to the new session, we skip it
	} else if existing.ACR > new.ACR {
		return nil
	}

	lifetime := time.Duration(new.RememberFor) * time.Second
	return as.sessions.Upsert(ctx, new, lifetime)
}

func (as *Service) GetSession(ctx context.Context, sessionID string) (Session, error) {
	return as.sessions.Get(ctx, sessionID)
}
