package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateSecretStorageBoxKeyShareTable() {
	goose.AddMigration(upCreateSecretStorageBoxKeyShareTable, downCreateSecretStorageBoxKeyShareTable)
}

func upCreateSecretStorageBoxKeyShareTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE secret_storage_box_key_share(
			id UUID PRIMARY KEY,
			invitation_share_hash VARCHAR(255) NOT NULL,
			encrypted_invitation_share VARCHAR(2047) NOT NULL,
			box_id UUID NOT NULL,
			account_root_key_hash VARCHAR(255) NOT NULL REFERENCES secret_storage_account_root_key (key_hash) ON DELETE CASCADE,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT one_per_box_per_root_key UNIQUE(box_id, account_root_key_hash)
		)
	`)
	if err != nil {
		return fmt.Errorf("creating table secret_storage_box_key_share: %v", err)
	}
	return nil
}

func downCreateSecretStorageBoxKeyShareTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE secret_storage_box_key_share;`); err != nil {
		return fmt.Errorf("dropping table secret_storage_box_key_share: %v", err)
	}
	return nil
}
