package events

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"

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
// Do handler is at least responsible for making the event persistent in storage (some events might do it differently tough).
// After handler must perform non-critical actions that might fail without altering the state of the box.
var eventTypeHandlerMapping = map[string]EventHandler{
	etype.Accessadd: {doAddAccess, nil},
	etype.Accessrm:  {doRmAccess, nil},

	etype.Memberleave: {doLeave, group(sendRealtimeUpdate, countActivity, invalidateCaches)},
	etype.Memberjoin:  {doJoin, group(sendRealtimeUpdate, countActivity, invalidateCaches)},

	etype.Msgdelete: {doDeleteMsg, group(sendRealtimeUpdate, computeUsedSpace)},
	etype.Msgedit:   {doEditMsg, group(sendRealtimeUpdate, computeUsedSpace)},
	etype.Msgfile:   {doMessage, group(sendRealtimeUpdate, countActivity, computeUsedSpace)},
	etype.Msgtext:   {doMessage, group(sendRealtimeUpdate, countActivity, computeUsedSpace)},

	etype.Stateaccessmode: {doStateAccessMode, group(sendRealtimeUpdate, countActivity)},
	etype.Statekeyshare:   {doStateKeyShare, nil},

	// never added by end-users directly but the system
	etype.Memberkick: {empty, group(notifyKick, sendRealtimeUpdate, countActivity, invalidateCaches)},
}

// group handlers declaration
func group(handlers ...afterHandler) []afterHandler {
	return handlers
}

// Handler ...
func Handler(eType string) EventHandler {
	return eventTypeHandlerMapping[eType]
}
