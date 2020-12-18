package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// Metadata ...
type Metadata interface{}

type doHandler func(
	ctx context.Context,
	event *Event,
	extraJSON null.JSON,
	exec boil.ContextExecutor, // transaction
	redConn *redis.Client,
	identityMapper *IdentityMapper,
	cryptoactions external.CryptoRepo,
	files files.FileStorageRepo,
) (Metadata, error)

type afterHandler func(
	context.Context,
	*Event,
	boil.ContextExecutor, // db connector
	*redis.Client,
	*IdentityMapper,
	files.FileStorageRepo,
	Metadata,
) error

func empty(_ context.Context, _ *Event, _ null.JSON, _ boil.ContextExecutor, _ *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	return nil, nil
}

// checkReferrer id is set
func checkReferrer(e Event) error {
	// check that the current sender has access to the box
	if err := v.ValidateStruct(&e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return err
	}

	return nil
}

// EventHandler ...
type EventHandler struct {
	Do    doHandler
	After []afterHandler
}

// NOTE:
// Do handler is at least responsible for making the event persistent in storage (some events might do it differently though).
// After handler must perform non-critical actions that might fail without altering the state of the box.
var eventTypeHandlerMapping = map[string]EventHandler{
	"state.key_share": {doKeyShare, nil},
	"msg.text":        {doMessage, gh(sendRealtimeUpdate, countActivity, computeUsedSpace)},
	"msg.file":        {doMessage, gh(sendRealtimeUpdate, countActivity, computeUsedSpace)},
	"msg.edit":        {doEditMsg, gh(sendRealtimeUpdate, computeUsedSpace)},
	"msg.delete":      {doDeleteMsg, gh(sendRealtimeUpdate, computeUsedSpace)},
	"access.add":      {doAddAccess, nil},
	"access.rm":       {doRmAccess, nil},
	"member.leave":    {doLeave, gh(sendRealtimeUpdate, countActivity, invalidateCaches)},
	"member.join":     {doJoin, gh(sendRealtimeUpdate, countActivity, invalidateCaches)},

	// never added by end-users directly but the system
	"member.kick": {empty, gh(notifyKick, sendRealtimeUpdate, countActivity, invalidateCaches)},
}

// group handlers declaration
func gh(handlers ...afterHandler) []afterHandler {
	return handlers
}

// Handler ...
func Handler(eType string) EventHandler {
	return eventTypeHandlerMapping[eType]
}
