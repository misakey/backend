package box

import (
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/config"
)

func initConfig() {
	// handle missing mandatory fields
	mandatoryFields := []string{
		"authflow.self_client_id",
		"hydra.admin_endpoint",
	}
	// no secret fields so far
	secretFields := []string{}
	config.FatalIfMissing("BOX", mandatoryFields)
	config.Print("BOX", secretFields)
}
