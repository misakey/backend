package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

type BoxInfo struct {
	ID          string
	Title       string
	NewMessages int
}

type DigestInfo struct {
	identity  identity.Identity
	boxesInfo []*BoxInfo
}

func (dj *DigestJob) SendDigests(ctx context.Context) error {

	logger.FromCtx(ctx).Info().Msgf("starting digests job with frequency %s", dj.frequency)

	digestInfos, identityIDs, err := dj.buildToNotifyInfo(ctx)
	if err != nil {
		return err
	}

	if len(identityIDs) == 0 {
		logger.FromCtx(ctx).Debug().Msg("nobody to send digests to")
		return nil
	}

	// get identities
	filters := identity.IdentityFilters{
		IDs: identityIDs,
	}
	users, err := dj.identities.List(ctx, filters)
	if err != nil {
		return err
	}

	// only keep identities for which configuration matches the current job
	// and not used by an active user recently
	for _, user := range users {
		dj.checkEligibility(ctx, *user, digestInfos)
	}

	// we keep box info in a cache to avoid too many calls to the db
	boxCache := make(map[string]events.Box)
	for id, digestInfo := range digestInfos {
		//get the boxes info (try to make profit of the already fetched information): title and silenced
		// get box settings
		var totalNewMessages int
		for _, boxInfo := range digestInfo.boxesInfo {
			boxID := boxInfo.ID
			_, ok := boxCache[boxID]
			if !ok {
				boxCache[boxID], err = events.Compute(ctx, boxID, dj.boxExec, dj.identityMapper, nil)
				if err != nil {
					logger.FromCtx(ctx).Error().Err(err).Msgf("could not get box %s", boxID)
					continue
				}
			}
			newMessages, err := dj.redConn.Get(cache.GetToNotifyKey(id, boxID)).Int()
			if err != nil {
				logger.FromCtx(ctx).Error().Err(err).Msgf("could not get new messages for box %s", boxID)
				continue
			}
			totalNewMessages += newMessages
			boxInfo.NewMessages = newMessages
			boxInfo.Title = boxCache[boxID].Title
		}

		//build and send the notification
		displayName := digestInfo.identity.DisplayName
		if len(displayName) > 24 {
			displayName = displayName[:20] + "..."
		}

		data := map[string]interface{}{
			"to":             digestInfo.identity.Identifier.Value,
			"displayName":    displayName,
			"firstLetter":    digestInfo.identity.DisplayName[:1],
			"avatarURL":      digestInfo.identity.AvatarURL.String,
			"boxes":          digestInfo.boxesInfo,
			"total":          totalNewMessages,
			"domain":         dj.domain,
			"accountBaseURL": fmt.Sprintf("https://%s/accounts/%s", dj.domain, id),
		}

		subject := "Misakey - Nouveau(x) messages"
		template := "notification"
		if digestInfo.identity.AccountID.IsZero() {
			template = "notificationNoAccount"
		}
		content, err := dj.templates.NewEmail(ctx, digestInfo.identity.Identifier.Value, subject, template, data)
		if err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not build email for %s", id)
			continue
		}

		logger.FromCtx(ctx).Debug().Msgf("sending email to %s", id)
		if err := dj.emails.Send(ctx, content); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not send email to %s", id)
			continue
		}

		//delete the keys toNotify:user_*
		if err := events.DelAllToNotifyForIdentity(ctx, dj.redConn, id); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not del toNotify key for user %s", id)
		}
	}

	return nil
}

// BuildToNotifyInfo build useful information to send users notifications
// It returns a map of digest info per user
// and a list of user ids to notify
func (dj *DigestJob) buildToNotifyInfo(ctx context.Context) (map[string]*DigestInfo, []string, error) {

	keys, err := events.GetAllToNotifyKeys(ctx, dj.redConn)
	if err != nil {
		return nil, nil, err
	}

	digestInfos := make(map[string]*DigestInfo)
	identityIDs := []string{}
	for _, key := range keys {
		elts := strings.Split(key, ":")
		userID := strings.Split(elts[1], "_")[1]
		boxID := elts[2]
		if _, ok := digestInfos[userID]; !ok {
			digestInfos[userID] = &DigestInfo{}
			identityIDs = append(identityIDs, userID)
		}
		digestInfos[userID].boxesInfo = append(digestInfos[userID].boxesInfo, &BoxInfo{ID: boxID})
	}

	return digestInfos, identityIDs, nil

}

// checkEligibility of identity for digests
// and remove identity from digestInfo if not eligible
// (warning: alters digestInfo)
func (dj *DigestJob) checkEligibility(ctx context.Context, identity identity.Identity, digestInfo map[string]*DigestInfo) {
	if _, ok := digestInfo[identity.ID]; !ok {
		// if the identity is not in the list, we don’t need to check its eligibility
		return
	}
	digestInfo[identity.ID].identity = identity

	// get identity last interaction with the app
	lastInteraction, err := dj.redConn.Get(fmt.Sprintf("lastInteraction:user_%s", identity.ID)).Int()
	if err != nil && err != redis.Nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("could not get last interaction for identity %s", identity.ID)
	}
	fromLastInteraction := time.Since(time.Unix(int64(lastInteraction), 0))

	// if the last interaction is sooner than the desired notification period
	// or the identity configuration does not match with the current job frequency
	// then do not send digest to the identity
	if fromLastInteraction < dj.period || identity.Notifications != dj.frequency {
		logger.FromCtx(ctx).Debug().Msgf("won’t send digest to %s", identity.ID)
		delete(digestInfo, identity.ID)
	}
}
