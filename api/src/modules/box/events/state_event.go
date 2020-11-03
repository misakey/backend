package events

import (
	"context"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type StateLifecycleContent struct {
	State string `json:"state"`
}

func doLifecycle(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, _ *redis.Client, _ *IdentityMapper, _ external.CryptoActionRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check accesses
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// handle content
	var c StateLifecycleContent
	if err := e.JSONContent.Unmarshal(&c); err != nil {
		return nil, merror.Transform(err).Describe("marshalling lifecycle content")
	}

	// referrer ID cannot be set
	if err := v.Empty.Validate(&e.ReferrerID); err != nil {
		return nil, err
	}
	// only closed state lifecycle change is allowed today
	if err := v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("closed")),
	); err != nil {
		return nil, err
	}
	return nil, e.persist(ctx, exec)
}

func notifyLifecycle(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	// TODO (perf): use metadahandlers to bear already retrieved data accross handlers ?
	box, err := Compute(ctx, e.BoxID, exec, identities, e)
	if err != nil {
		return merror.Transform(err).Describe("computing box")
	}

	// notify the member the lifecychel has changed.
	kickDetails := struct {
		BoxID     string `json:"id"`
		BoxTitle  string `json:"title"`
		Lifecycle string `json:"lifecycle"`
	}{
		BoxID:     box.ID,
		BoxTitle:  box.Title,
		Lifecycle: box.Lifecycle,
	}
	bytes, err := json.Marshal(kickDetails)
	if err != nil {
		return merror.Transform(err).Describe("marshalling kick details")
	}
	details := null.JSONFrom(bytes)

	// retrieve member ids to notify
	memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return merror.Transform(err).Describe("notifying member: listing members")
	}
	// skip the event sender id
	ids := memberIDs[:0]
	for _, identityID := range memberIDs {
		if identityID == e.SenderID {
			continue
		}
		ids = append(ids, identityID)
	}
	identities.CreateNotifs(ctx, ids, "box.lifecycle", details)
	return nil
}

func isClosed(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) (bool, error) {
	jsonQuery := `{"state": "closed"}`
	_, err := get(ctx, exec, eventFilters{
		boxID:   null.StringFrom(boxID),
		eType:   null.StringFrom("state.lifecycle"),
		content: &jsonQuery,
	})
	if err != nil {
		if merror.HasCode(err, merror.NotFoundCode) {
			return false, nil
		}
		return true, merror.Transform(err).Describe("getting closed lifecycle")
	}
	return true, nil
}

func MustBoxBeOpen(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	closed, err := isClosed(ctx, exec, boxID)
	if err != nil {
		return err
	}
	if closed {
		return merror.Conflict().Describe("box is closed").
			Detail("lifecycle", merror.DVConflict)
	}
	return nil
}
