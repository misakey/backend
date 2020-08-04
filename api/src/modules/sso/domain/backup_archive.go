package domain

import (
	"time"

	"github.com/volatiletech/null"
)

type BackupArchive struct {
	ID          string
	AccountID   string
	Data        null.String
	CreatedAt   time.Time
	RecoveredAt null.Time
	DeletedAt   null.Time
}
