package authn

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
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
	return srr.SimpleKeyRedis.Set(ctx, srr.key(session.ID), []byte(session.ACR), lifetime)
}

func (srr SessionRedisRepo) Get(ctx context.Context, sessionID string) (Session, error) {
	session := Session{
		ID: sessionID,
	}
	secLevelBytes, err := srr.SimpleKeyRedis.Get(ctx, srr.key(sessionID))
	if err != nil {
		return session, err
	}
	session.ACR = oidc.ClassRef(secLevelBytes)
	return session, nil
}
