package authn

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
)

func (as *Service) UpsertSession(ctx context.Context, new authn.Session) error {
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

func (as *Service) GetSession(ctx context.Context, sessionID string) (authn.Session, error) {
	return as.sessions.Get(ctx, sessionID)
}
