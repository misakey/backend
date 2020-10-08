package oidc

import (
	customclock "github.com/benbjohnson/clock"
)

// declare clock - time by default
// this variable is used for mocking purpose
var clock = customclock.New()
