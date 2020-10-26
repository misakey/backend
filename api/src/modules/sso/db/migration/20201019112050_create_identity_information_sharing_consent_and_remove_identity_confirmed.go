package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed() {
	goose.AddMigration(upCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed, downCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed)
}

func upCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed(tx *sql.Tx) error {
	// create identity_profile_sharing_consent
	// NOTE: require index on identity_id: has been done in the next migration
	if _, err := tx.Exec(`
		CREATE TABLE identity_profile_sharing_consent(
			id SERIAL PRIMARY KEY,
			identity_id UUID  NOT NULL REFERENCES identity ON DELETE CASCADE,
			information_type VARCHAR(32) NOT NULL,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			revoked_at timestamptz DEFAULT NULL
		)
	`); err != nil {
		return fmt.Errorf("creating identity_profile_sharing_consent: %v", err)
	}
	// remove identity confirmed column that is unused - not planned to be used
	if _, err := tx.Exec(`ALTER TABLE identity
		DROP COLUMN confirmed;
	`); err != nil {
		return fmt.Errorf("removing confirmed column: %v", err)
	}
	return nil
}

func downCreateIdentityProfileSharingConsentAndRemoveIdentityConfirmed(tx *sql.Tx) error {
	// drop the created table
	if _, err := tx.Exec(`DROP TABLE identity_profile_sharing_consent;`); err != nil {
		return fmt.Errorf("dropping identity_profile_sharing_consent: %v", err)
	}
	// re-add confirmed boolean to identity and set everything to true (unused anyway)
	_, err := tx.Exec(`ALTER TABLE identity
		ADD COLUMN confirmed BOOLEAN NOT NULL DEFAULT true;
	`)
	if err != nil {
		return fmt.Errorf("adding back confirmed column: %v", err)
	}
	return nil
}
