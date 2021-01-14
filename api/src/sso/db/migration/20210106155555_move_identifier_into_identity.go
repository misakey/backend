package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

func initMoveIdentifierIntoIdentity() {
	goose.AddMigration(UpMoveIdentifierIntoIdentity, DownMoveIdentifierIntoIdentity)
}

func UpMoveIdentifierIntoIdentity(tx *sql.Tx) error {
	// 1. create columns on identity for identifier
	// alter table
	_, err := tx.Exec(`ALTER TABLE identity
		ADD COLUMN identifier_kind VARCHAR(32),
		ADD COLUMN identifier_value VARCHAR(255) UNIQUE,
		DROP COLUMN is_authable;
	`)
	if err != nil {
		return merr.From(err).Desc("create identity.identifier_*")
	}

	// 2. copy column value from identifier to their respective identity
	_, err = tx.Exec(`
		UPDATE identity SET
			identifier_value = identifier.value,
			identifier_kind = identifier.kind
		FROM identifier WHERE identifier.id = identity.identifier_id;
	`)
	if err != nil {
		return merr.From(err).Desc("copying value from identifier to identity")
	}
	_, err = tx.Exec(`ALTER TABLE identity
		ALTER COLUMN identifier_kind SET NOT NULL,
		ALTER COLUMN identifier_value SET NOT NULL;
	`)
	if err != nil {
		return merr.From(err).Desc("alter nullability of identity.identifier_kind/value")
	}

	// 3. remove identifier_id column on identity table
	_, err = tx.Exec(`ALTER TABLE identity DROP COLUMN identifier_id;`)
	if err != nil {
		return merr.From(err).Desc("drop identity.identifier_id")
	}

	// 4. remove identifier table
	_, err = tx.Exec(`DROP TABLE identifier;`)
	if err != nil {
		return merr.From(err).Desc("drop identifier table")
	}
	return nil
}

func DownMoveIdentifierIntoIdentity(tx *sql.Tx) error {
	// 1. create identifier table
	// create table
	// NOTE: do not recreate the idx onvaleu since UNIQUE constraint already build one
	_, err := tx.Exec(`
		CREATE TABLE identifier(
			id UUID PRIMARY KEY,
			kind VARCHAR(32) NOT NULL,
			value VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return merr.From(err).Desc("create identifier table")
	}

	// 2. create identifier_id column on identity table
	_, err = tx.Exec(`ALTER TABLE identity
		ADD COLUMN identifier_id UUID DEFAULT NULL REFERENCES identifier,
		ADD is_authable BOOLEAN NOT NULL DEFAULT TRUE;
	`)
	if err != nil {
		return merr.From(err).Desc("create identity.identifier_id")
	}

	// 3. create identifier from identity
	// md5(random()::text || clock_timestamp()::text)::uuid is the way to natively generate an UUID
	_, err = tx.Exec(`INSERT INTO identifier(id, kind, value) SELECT md5(random()::text || clock_timestamp()::text)::uuid, identity.identifier_kind, identity.identifier_value FROM identity;`)
	if err != nil {
		return merr.From(err).Desc("create identifier from identity")
	}
	_, err = tx.Exec(`UPDATE identity SET identifier_id = identifier.id FROM identifier WHERE identifier.value = identity.identifier_value;`)
	if err != nil {
		return merr.From(err).Desc("bind identity to create identifiers")
	}

	// 4. Set identity.identifier_id not nullable
	_, err = tx.Exec(`ALTER TABLE identity
		ALTER COLUMN identifier_id SET NOT NULL;
	`)
	if err != nil {
		return merr.From(err).Desc("alter nullability of identity.identifier_id")
	}
	_, err = tx.Exec(`CREATE UNIQUE INDEX identity_authable_identifier_idx ON identity (identifier_id, is_authable);`)
	if err != nil {
		return merr.From(err).Desc("create index identity_authable_identifier_idx")
	}
	_, err = tx.Exec(`CREATE UNIQUE INDEX identity_identifier_account_idx
		ON identity (account_id, identifier_id);`)
	if err != nil {
		return err
	}

	// 5. remove identity identifier_* column
	_, err = tx.Exec(`ALTER TABLE identity
		DROP COLUMN identifier_value,
		DROP COLUMN identifier_kind;
	`)
	if err != nil {
		return merr.From(err).Desc("drop identity.identifier_*")
	}
	return nil
}
