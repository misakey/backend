package csrf

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

func GenerateToken(accessToken string, expiration time.Duration, redConn *redis.Client) (string, error) {
	uuidToken, err := uuid.NewString()
	if err != nil {
		return "", err
	}

	token := strings.ReplaceAll(uuidToken, "-", "")
	if _, err := redConn.Set(fmt.Sprintf("csrf:%s", accessToken), token, expiration).Result(); err != nil {
		return "", err
	}

	return token, nil
}

func IsTokenValid(accessToken, csrfToken string, redConn *redis.Client) bool {
	storedCsrfToken, err := redConn.Get(fmt.Sprintf("csrf:%s", accessToken)).Result()
	if err != nil {
		return false
	}

	if storedCsrfToken == csrfToken {
		return true
	}

	return false
}
