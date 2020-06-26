package format

import "regexp"

var UnpaddedURLSafeBase64 = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
