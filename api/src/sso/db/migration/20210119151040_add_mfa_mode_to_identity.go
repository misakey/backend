package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

func initAddMFAMethodToIdentity() {
	goose.AddMigration(UpAddMFAMethodToIdentity, DownAddMFAMethodToIdentity)
}

func UpAddMFAMethodToIdentity(tx *sql.Tx) error {
	// 1. create the column on identity for MFA_method
	// alter table
	_, err := tx.Exec(`
		ALTER TABLE identity
		ADD COLUMN mfa_method VARCHAR(32) NOT NULL DEFAULT 'disabled';
	`)
	if err != nil {
		return merr.From(err).Desc("create identity.mfa_method")
	}
	return nil
}

func DownAddMFAMethodToIdentity(tx *sql.Tx) error {
	// 1. drop the column on identity for MFA_method
	// alter table
	_, err := tx.Exec(`
		ALTER TABLE identity
		DROP COLUMN mfa_method;
	`)
	if err != nil {
		return merr.From(err).Desc("drop identity.mfa_method")
	}
	return nil
}
