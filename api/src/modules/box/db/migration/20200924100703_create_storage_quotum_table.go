package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func initCreateStorageQuotumTable() {
	goose.AddMigration(upCreateStorageQuotumTable, downCreateStorageQuotumTable)
}

func upCreateStorageQuotumTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE storage_quotum(
		id UUID PRIMARY KEY,
		identity_id UUID NOT NULL,
		value BIGINT NOT NULL,
		origin VARCHAR(255) NOT NULL,
		created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	return err
}

func downCreateStorageQuotumTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE storage_quotum;`)
	return err
}
