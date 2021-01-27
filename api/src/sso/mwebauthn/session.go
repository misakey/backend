package mwebauthn

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/duo-labs/webauthn/webauthn"
	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// StoreSession in redis
func StoreSession(redConn *redis.Client, sessionData *webauthn.SessionData, identityID, challenge string) error {
	// store session data in redis
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	_, err = redConn.Set(fmt.Sprintf("%s:%s:webauthn", identityID, challenge), sessionDataJSON, 360*time.Second).Result()

	return err
}

// GetSession from redis
func GetSession(redConn *redis.Client, identityID, challenge string) (webauthn.SessionData, error) {
	var sessionData webauthn.SessionData
	sessionDataJSON, err := redConn.Get(fmt.Sprintf("%s:%s:webauthn", identityID, challenge)).Result()
	if err != nil {
		return sessionData, merr.From(err).Desc("getting session")
	}
	if err := json.Unmarshal([]byte(sessionDataJSON), &sessionData); err != nil {
		return sessionData, merr.From(err).Desc("decoding session data")
	}

	return sessionData, nil
}
