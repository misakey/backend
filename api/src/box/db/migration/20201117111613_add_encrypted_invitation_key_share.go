package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func initAddEncryptedInvitationKeyShare() {
	goose.AddMigration(upAddEncryptedInvitationKeyShare, downAddEncryptedInvitationKeyShare)
}

func upAddEncryptedInvitationKeyShare(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE box_key_share
			ADD COLUMN encrypted_invitation_key_share VARCHAR(1023);
	`)
	return err
}

func downAddEncryptedInvitationKeyShare(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE box_key_share
			DROP COLUMN encrypted_invitation_key_share;
	`)
	return err
}
