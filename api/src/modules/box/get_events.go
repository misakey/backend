package box

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func (h *handler) listEvents(ctx echo.Context) error {
	// TODO access control

	boxID := ctx.Param("id")
	err := validation.Validate(boxID, validation.Required, is.UUIDv4)
	if err != nil {
		return merror.Transform(err).Code(merror.BadRequestCode).From(merror.OriPath)
	}

	// list
	boxEvents, err := events.List(ctx.Request().Context(), boxID, h.db)
	if err != nil {
		return err
	}

	sendersMap, err := mapSenderIdentities(ctx.Request().Context(), boxEvents, h.identityRepo)
	if err != nil {
		return merror.Transform(err).Describe("retrieving events senders")
	}

	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		sender := sendersMap[e.SenderID]
		views[i] = events.ToView(e, sender)
	}

	return ctx.JSON(http.StatusOK, views)
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

	identities, err := identityRepo.ListIdentities(ctx, domain.IdentityFilters{IDs: senderIDs})
	if err != nil {
		return nil, err
	}

	var sendersMap = make(map[string]domain.Identity, len(identities))
	for _, identity := range identities {
		sendersMap[identity.ID] = *identity
	}

	return sendersMap, nil
}
