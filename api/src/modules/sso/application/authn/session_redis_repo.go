package authn

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
)

type AuthnSessionRedis struct {
	repositories.SimpleKeyRedis
}

func NewAuthSessionRedis(redConn *redis.Client) AuthnSessionRedis {
	return AuthnSessionRedis{repositories.NewSimpleKeyRedis(redConn)}
}

func (ars AuthnSessionRedis) key(sessionID string) string {
	return "authn_session:" + sessionID
}

func (ars AuthnSessionRedis) Upsert(ctx context.Context, session authn.Session, lifetime time.Duration) error {
	return ars.SimpleKeyRedis.Set(ctx, ars.key(session.ID), []byte(session.ACR), lifetime)
}

func (ars AuthnSessionRedis) Get(ctx context.Context, sessionID string) (authn.Session, error) {
	session := authn.Session{
		ID: sessionID,
	}
	secLevelBytes, err := ars.SimpleKeyRedis.Get(ctx, ars.key(sessionID))
	if err != nil {
		return session, err
	}
	session.ACR = authn.ClassRef(secLevelBytes)
	return session, nil
}
