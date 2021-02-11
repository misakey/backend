package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"
)

func initCreateSecretStorageAccountRootKeyTable() {
	goose.AddMigration(upCreateSecretStorageAccountRootKeyTable, downCreateSecretStorageAccountRootKeyTable)
}

func upCreateSecretStorageAccountRootKeyTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE secret_storage_account_root_key(
			key_hash VARCHAR(255) PRIMARY KEY,
			account_id UUID NOT NULL REFERENCES account ON DELETE CASCADE,
			encrypted_key VARCHAR(1023) NOT NULL,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("creating table secret_storage_account_root_key: %v", err)
	}
	return nil
}

func downCreateSecretStorageAccountRootKeyTable(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE secret_storage_account_root_key;`); err != nil {
		return fmt.Errorf("dropping table secret_storage_account_root_key: %v", err)
	}
	return nil
}
