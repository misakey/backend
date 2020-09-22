package db

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/pressly/goose"
)

// StartMigration use goose to run imported migrations
func StartMigration(dsn string, migrationDir string) {
	// define command - default is up.
	command := "up"
	commandArg := false
	if len(os.Args) > 1 {
		// we look for a `--goose=` argument to retrieve the command
		for _, argument := range os.Args {
			if strings.Contains(argument, "--goose=") {
				command = strings.Trim(strings.Split(argument, "=")[1], "\"")
				log.Printf("command `%s` detected.", command)
				commandArg = true
				break
			}
		}
		if !commandArg {
			log.Println("no command detected, consider `up` as the default one.")
		}
	} else {
		log.Println("no command detected, consider `up` as the default one.")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("could not open db (%v)\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping db (%v)\n", err)
	}

	if err := goose.Run(command, db, migrationDir, command); err != nil {
		log.Fatalf("could not run goose (%v : %v)\n", command, err)
	}
}
