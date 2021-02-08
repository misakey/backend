package mrand

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StringWithCharset(length int, charset string) (string, error) {
	b := make([]byte, length)
	for i := range b {
		nb, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[nb.Int64()]
	}
	return string(b), nil
}

func String(length int) (string, error) {
	return StringWithCharset(length, charset)
}
