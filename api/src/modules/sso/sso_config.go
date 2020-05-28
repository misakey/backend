package sso

import (
	"github.com/spf13/viper"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/config"
)

func initConfig() {
	// set defaults value for configuration
	viper.SetDefault("hydra.secure", true)

	// handle missing mandatory fields
	mandatoryFields := []string{
		"authflow.self_client_id",
		"authflow.hydra_token_url",
		"authflow.self_encoded_jwk",
		"hydra.public_endpoint",
		"hydra.admin_endpoint",
		"authflow.auth_url",
		"authflow.code_redirect_url",
		"authflow.hydra_token_url",
		"authflow.token_redirect_url",
	}
	config.FatalIfMissing("SSO", mandatoryFields)
	secretFields := []string{
		"authflow.self_encoded_jwk",
	}
	config.Print("SSO", secretFields)
}
