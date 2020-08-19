package domain

import (
	"time"

	"github.com/volatiletech/null"
)

type BackupArchive struct {
	ID          string      `json:"id"`
	AccountID   string      `json:"account_id"`
	Data        null.String `json:"-"`
	CreatedAt   time.Time   `json:"created_at"`
	RecoveredAt null.Time   `json:"recovered_at"`
	DeletedAt   null.Time   `json:"deleted_at"`
}
