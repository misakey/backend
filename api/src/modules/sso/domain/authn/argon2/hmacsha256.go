package argon2

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

const hmacSaltLength = 16

func generateRandomSalt() ([]byte, error) {
	b := make([]byte, hmacSaltLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func hash(data []byte, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}
