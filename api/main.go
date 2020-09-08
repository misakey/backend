package main

import (
	"log"
	"os"

	"gitlab.misakey.dev/misakey/backend/api/src/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
