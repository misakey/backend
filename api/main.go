package main

import (
	"fmt"
	"os"

	"gitlab.misakey.dev/misakey/backend/api/src/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
