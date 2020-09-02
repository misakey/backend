package migration

import (
	"os"

	_ "github.com/lib/pq"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/migration"
)

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

	migration.StartGoose(os.Getenv("DSN_BOX"), os.Getenv("MIGRATION_DIR_BOX"))
}
