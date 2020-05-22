package migration

import (
	"database/sql"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upCreateAccountTable, downCreateAccountTable)
}

func upCreateAccountTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE account(
          id UUID PRIMARY KEY,
          password VARCHAR (255) NOT NULL
      );`)
	return err
}

func downCreateAccountTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE account;`)
	return err
}
