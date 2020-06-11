package boxes

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/repositories/sqlboiler"
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

	dbEvents, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt),
	).All(ctx.Request().Context(), h.DB)
	if err != nil {
		return merror.Transform(err).Describe("retrieving events")
	}

	if dbEvents == nil {
		return merror.NotFound().Describef("no box with id %s", boxID)
	}

	sendersMap, err := mapSenderIdentities(ctx.Request().Context(), dbEvents, h.IdentityService)
	if err != nil {
		return merror.Transform(err).Describe("retrieving events senders")
	}

	var result []events.View
	for _, e := range dbEvents {
		sender := sendersMap[e.SenderID]
		result = append(result, events.ToView(events.FromSqlBoiler(e), sender))
	}

	return ctx.JSON(http.StatusOK, result)
}

func mapSenderIdentities(ctx context.Context, dbEvents sqlboiler.EventSlice, identityService entrypoints.IdentityIntraprocessInterface) (map[string]domain.Identity, error) {
	// getting senders IDs without duplicates
	var senderIDs []string
	idMap := make(map[string]bool)
	for _, event := range dbEvents {
		_, alreadyPresent := idMap[event.SenderID]
		if !alreadyPresent {
			senderIDs = append(senderIDs, event.SenderID)
			idMap[event.SenderID] = true
		}
	}

	identities, err := identityService.ListIdentities(ctx, domain.IdentityFilters{IDs: senderIDs})
	if err != nil {
		return nil, err
	}

	var sendersMap = make(map[string]domain.Identity, len(identities))
	for _, identity := range identities {
		sendersMap[identity.ID] = *identity
	}

	return sendersMap, nil
}
