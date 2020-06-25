package box

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func (h *handler) listEvents(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	boxID := eCtx.Param("id")
	err := validation.Validate(boxID, validation.Required, is.UUIDv4)
	if err != nil {
		return merror.Transform(err).Code(merror.BadRequestCode).From(merror.OriPath)
	}

	// if the box is closed, only the creator can list its events
	if err := boxes.MustBeCreatorIfClosed(ctx, h.repo.DB(), boxID, acc.Subject); err != nil {
		return err
	}

	// list
	boxEvents, err := events.ListByBoxID(ctx, h.repo.DB(), boxID)
	if err != nil {
		return err
	}

	sendersMap, err := mapSenderIdentities(ctx, boxEvents, h.repo.Identities())
	if err != nil {
		return merror.Transform(err).Describe("retrieving events senders")
	}

	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		sender := sendersMap[e.SenderID]
		views[i] = events.ToView(e, sender)
	}

	return eCtx.JSON(http.StatusOK, views)
}

func mapSenderIdentities(ctx context.Context, events []events.Event, identityRepo entrypoints.IdentityIntraprocessInterface) (map[string]domain.Identity, error) {
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
