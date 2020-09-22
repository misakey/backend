package logger

import (
	"os"
	"path/filepath"
)

// GetEchoFormat returns echo format services should use.
func GetEchoFormat() string {
	serviceName := filepath.Base(os.Args[0])
	serviceVersion := os.Getenv("VERSION")
	if os.Getenv("ENV") == "development" {
		return "time=${time_rfc3339}, request_id=${id}, fapi_interaction_id=${header:X-Fapi-Interaction-Id}, method=${method}, uri=${uri}, status=${status}\n"
	}
	return `{"time": "${time_rfc3339}", ` +
		`"service_name": "` + serviceName + `", ` +
		`"service_version": "` + serviceVersion + `", ` +
		`"request_id": "${id}", ` +
		`"fapi_interaction_id": "${header:x-fapi-interaction-id}",` +
		`"protocol": "http", ` +
		`"uri_path": "${path}", ` +
		`"method": "${method}", ` +
		`"request_bytes": ${bytes_in}, ` +
		`"status": ${status}, ` +
		`"response_bytes": ${bytes_out}, ` +
		`"response_time": "${latency}"}` +
		"\n"
}
