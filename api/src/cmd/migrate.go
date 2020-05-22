package cmd

import (
	"github.com/spf13/cobra"

	_ "github.com/lib/pq"

	"gitlab.misakey.dev/misakey/msk-sdk-go/migration"

	_ "gitlab.misakey.dev/misakey/backend/api/src/modules/sso/db/migration"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the postgreSQL Database",
	Long:  `A Goose migrator to migrate a postgreSQL Database in go.`,
	Run: func(cmd *cobra.Command, args []string) {
		migration.StartGoose()
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
