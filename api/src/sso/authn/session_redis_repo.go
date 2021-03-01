package authn

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mredis"
)

// SessionRedisRepo ...
type SessionRedisRepo struct {
	mredis.SimpleKeyRedis
}

// NewAuthnSessionRedis ...
func NewAuthnSessionRedis(skr mredis.SimpleKeyRedis) SessionRedisRepo {
	return SessionRedisRepo{skr}
}

func (srr SessionRedisRepo) key(sessionID string) string {
	return "authn_session:" + sessionID
}

// Upsert ...
func (srr SessionRedisRepo) Upsert(ctx context.Context, session Session, lifetime time.Duration) error {
	value, err := json.Marshal(session)
	if err != nil {
		return merr.From(err).Desc("marshaling sesion")
	}
	return srr.SimpleKeyRedis.Set(ctx, srr.key(session.ID), value, lifetime)
}

// Get ...
func (srr SessionRedisRepo) Get(ctx context.Context, sessionID string) (Session, error) {
	session := Session{}

	value, err := srr.SimpleKeyRedis.Get(ctx, srr.key(sessionID))
	if err != nil {
		return session, err
	}
	if err := json.Unmarshal(value, &session); err != nil {
		return session, merr.From(err).Desc("unmarshaling session")
	}
	session.ID = sessionID
	return session, nil
}
