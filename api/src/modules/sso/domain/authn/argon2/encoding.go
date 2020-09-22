package argon2

import (
	"encoding/base64"
	"fmt"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

const (
	algorithmID = "com.misakey.argon2-relief-v1"
)

func encode(argon2Params Params, serverSalt []byte, finalHash []byte) string {
	// XXX What about just using JSON or MsgPack for encoding/decoding?
	// this would probably require to increase the size limit of the "password" DB column though
	// (currently it's "VARCHAR(255)")
	return fmt.Sprintf(
		"$%s$m=%d,t=%d,p=%d$%s$%s$%s",
		algorithmID,
		argon2Params.Memory,
		argon2Params.Iterations,
		argon2Params.Parallelism,
		argon2Params.SaltBase64,
		base64.RawStdEncoding.EncodeToString(serverSalt),
		base64.RawStdEncoding.EncodeToString(finalHash),
	)
}

func decode(encodedHash string) (params Params, hmacSalt []byte, finalHash []byte, err error) {
	parts := strings.Split(encodedHash, "$")

	// because the string *starts* with a "$",
	// there should be 6 parts
	// with "part[0]" being the empty string
	if len(parts) != 6 || parts[1] != algorithmID {
		err = merror.Forbidden().Describef(`the stored hash is not in format "%s"`, algorithmID)
		return
	}

	_, err = fmt.Sscanf(
		parts[2], "m=%d,t=%d,p=%d",
		&params.Memory, &params.Iterations, &params.Parallelism,
	)
	if err != nil {
		return
	}

	// no need to decode this one because the server does not process it
	// it just sends it to the client
	params.SaltBase64 = parts[3]

	hmacSalt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return
	}

	finalHash, err = base64.RawStdEncoding.DecodeString(parts[5])
	return
}
