package migration

import (
	"os"

	_ "github.com/lib/pq" // import psql driver
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/db"
)

// Launch the migration
func Launch() {
	// /!\ Do not forget to add you new migration here
	initAddEventsTable()
	initCreateKeyShareTable()
	initAddBoxIDColumnToKeyShareTable()
	initRenameKeyShareTable()
	initCreateEncryptedFileTable()
	initCreateSavedFileTable()
	initCreateUniqueIndexOnSavedFileTable()
	initAddReferColumnOnBoxEvents()
	initCreateStorageQuotumTable()
	initCreateBoxUsedSpaceTable()
	initCreateBoxSettingsTable()
	initAddEncryptedInvitationKeyShare()

	db.StartMigration(os.Getenv("DSN_BOX"), os.Getenv("MIGRATION_DIR_BOX"))
}
