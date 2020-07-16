package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	ssoentrypoints "gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type ListEventsRequest struct {
	boxID string
}

func (req *ListEventsRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, is.UUIDv4),
	)
}

func (bs *BoxApplication) ListEvents(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ListEventsRequest)
	acc := ajwt.GetAccesses(ctx)

	// if the box is closed, only the creator can list its events
	if err := boxes.MustBeCreatorIfClosed(ctx, bs.db, req.boxID, acc.Subject); err != nil {
		return nil, err
	}

	// list
	boxEvents, err := events.ListByBoxID(ctx, bs.db, req.boxID)
	if err != nil {
		return nil, err
	}

	sendersMap, err := mapSenderIdentities(ctx, boxEvents, bs.identities)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving events senders")
	}

	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		sender := sendersMap[e.SenderID]
		views[i] = events.ToView(e, sender)
	}

	return views, nil
}

func mapSenderIdentities(ctx context.Context, events []events.Event, identityRepo ssoentrypoints.IdentityIntraprocessInterface) (map[string]domain.Identity, error) {
	// getting senders IDs without duplicates
	var senderIDs []string
	idMap := make(map[string]bool)
	for _, event := range events {
		_, alreadyPresent := idMap[event.SenderID]
		if !alreadyPresent {
			senderIDs = append(senderIDs, event.SenderID)
			idMap[event.SenderID] = true
		}
	}

	identities, err := identityRepo.List(ctx, domain.IdentityFilters{IDs: senderIDs})
	if err != nil {
		return nil, err
	}

	var sendersMap = make(map[string]domain.Identity, len(identities))
	for _, identity := range identities {
		sendersMap[identity.ID] = *identity
	}

	return sendersMap, nil
}
