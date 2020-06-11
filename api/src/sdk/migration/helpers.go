package migration

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/pressly/goose"
)

// StartGoose use goose to run imported migrations
func StartGoose(dsn string) {
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

	// configure dir where migration belongs
	dir := os.Getenv("MIGRATION_DIR")
	if len(dir) == 0 {
		dir = "./"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("could not open db (%v)\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping db (%v)\n", err)
	}

	if err := goose.Run(command, db, dir, command); err != nil {
		log.Fatalf("could not run goose (%v : %v)\n", command, err)
	}
}
