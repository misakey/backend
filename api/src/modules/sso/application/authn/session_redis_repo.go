package authn

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type SessionRedisRepo struct {
	repositories.SimpleKeyRedis
}

func NewAuthnSessionRedis(redConn *redis.Client) SessionRedisRepo {
	return SessionRedisRepo{repositories.NewSimpleKeyRedis(redConn)}
}

func (srr SessionRedisRepo) key(sessionID string) string {
	return "authn_session:" + sessionID
}

func (srr SessionRedisRepo) Upsert(ctx context.Context, session Session, lifetime time.Duration) error {
	value, err := json.Marshal(session)
	if err != nil {
		return merror.Transform(err).Describe("marshaling sesion")
	}
	return srr.SimpleKeyRedis.Set(ctx, srr.key(session.ID), value, lifetime)
}

func (srr SessionRedisRepo) Get(ctx context.Context, sessionID string) (Session, error) {
	session := Session{}

	value, err := srr.SimpleKeyRedis.Get(ctx, srr.key(sessionID))
	if err != nil {
		return session, err
	}
	if err := json.Unmarshal(value, &session); err != nil {
		return session, merror.Transform(err).Describe("unmarshaling session")
	}
	session.ID = sessionID
	return session, nil
}
