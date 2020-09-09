package domain

import (
	"time"

	"github.com/volatiletech/null"
)

type CryptoAction struct {
	ID                  string      `json:"id"`
	AccountID           string      `json:"-"`
	SenderIdentityID    null.String `json:"-"`
	Type                string      `json:"type"`
	BoxID               null.String `json:"box_id"`
	EncryptionPublicKey string      `json:"encryption_public_key"`
	Encrypted           string      `json:"encrypted"`
	CreatedAt           time.Time   `json:"created_at"`
}
