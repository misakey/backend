package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// BoxInfo model
type BoxInfo struct {
	ID          string
	Title       string
	NewMessages int
}

// DigestInfo model
type DigestInfo struct {
	identity  identity.Identity
	boxesInfo []*BoxInfo
}

// SendDigests checks users to notify
// and send them their digests
func (dj *DigestJob) SendDigests(ctx context.Context) error {

	logger.FromCtx(ctx).Info().Msgf("starting digests job with frequency %s", dj.frequency)

	digestInfos, identityIDs, err := dj.buildDigestCountInfo(ctx)
	if err != nil {
		return err
	}

	if len(identityIDs) == 0 {
		logger.FromCtx(ctx).Debug().Msg("nobody to send digests to")
		return nil
	}

	// check eligibility
	// get first identities for the notification configuration linked to it
	filters := identity.Filters{
		IDs:            identityIDs,
		IdentifierKind: null.StringFrom(string(identity.IdentifierKindEmail)),
	}
	identities, err := identity.List(ctx, dj.ssoDB, filters)
	if err != nil {
		return err
	}
	for _, identity := range identities {
		// remove non-eligible identity from digestInfos map
		if ok := dj.isEligible(ctx, *identity); !ok {
			logger.FromCtx(ctx).Debug().Msgf("won’t send digest to %s", identity.ID)
			delete(digestInfos, identity.ID)
		} else { // otherwise bind it to the digest info
			digestInfos[identity.ID].identity = *identity
		}
	}

	// we keep box title in a cache to avoid too many calls to the db
	boxTitleCache := make(map[string]string)
	for userID, digestInfo := range digestInfos {
		//get the boxes info (try to make profit of the already fetched information): title and silenced
		// get box settings
		var totalNewMessages int
		for _, boxInfo := range digestInfo.boxesInfo {
			boxID := boxInfo.ID
			_, ok := boxTitleCache[boxID]
			if !ok {
				boxInfo, err := events.GetCreateInfo(ctx, dj.boxDB, boxID)
				if err != nil {
					logger.FromCtx(ctx).Error().Err(err).Msgf("could not get box %s title", boxID)
					continue
				}
				boxTitleCache[boxID] = boxInfo.Title
			}
			newMsgCount, err := dj.redConn.Get(cache.DigestCountKeyByUserBox(userID, boxID)).Int()
			if err != nil {
				logger.FromCtx(ctx).Error().Err(err).Msgf("could not get new messages count for box %s", boxID)
				continue
			}
			totalNewMessages += newMsgCount
			boxInfo.NewMessages = newMsgCount
			boxInfo.Title = boxTitleCache[boxID]
		}

		// build and send the notification
		displayName := digestInfo.identity.DisplayName
		if len(displayName) > 24 {
			displayName = displayName[:20] + "..."
		}

		data := map[string]interface{}{
			"to":             digestInfo.identity.IdentifierValue,
			"displayName":    displayName,
			"firstLetter":    digestInfo.identity.DisplayName[:1],
			"avatarURL":      digestInfo.identity.AvatarURL.String,
			"boxes":          digestInfo.boxesInfo,
			"total":          totalNewMessages,
			"domain":         dj.domain,
			"accountBaseURL": fmt.Sprintf("https://%s/accounts/%s", dj.domain, userID),
		}

		subject := "Misakey - Nouveau(x) message(s)"
		template := "notification"
		if digestInfo.identity.AccountID.IsZero() {
			template = "notificationNoAccount"
		}
		content, err := dj.templates.NewEmail(ctx, digestInfo.identity.IdentifierValue, subject, template, data)
		if err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not build email for %s", userID)
			continue
		}

		logger.FromCtx(ctx).Debug().Msgf("sending email to %s", userID)
		if err := dj.emails.Send(ctx, content); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not send email to %s", userID)
			continue
		}

		//delete the keys digestCount:user_*
		if err := events.DelAllDigestCountForIdentity(ctx, dj.redConn, userID); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msgf("could not del digestCount key for user %s", userID)
		}
	}

	return nil
}

// BuildDigestCountInfo build useful information to send users notifications
// It returns a map of digest info per user
// and a list of user ids to notify
func (dj *DigestJob) buildDigestCountInfo(ctx context.Context) (map[string]*DigestInfo, []string, error) {
	keys, err := events.GetAllDigestCountKeys(ctx, dj.redConn)
	if err != nil {
		return nil, nil, err
	}

	digestInfos := make(map[string]*DigestInfo)
	identityIDs := []string{}
	for _, key := range keys {
		elts := strings.Split(key, ":")
		userID := strings.Split(elts[1], "_")[1]
		boxPart := strings.Split(elts[2], "_")
		// NOTE: previously the box part was the id of the box directly
		// with commit 403b8c31f30d0a0c4e1381ddf3ea7c691dca67e0 it is now box_{id}
		// the system has to handle previous keys format for a while
		boxID := boxPart[0]
		if len(boxPart) == 2 {
			boxID = boxPart[1]
		}
		if _, ok := digestInfos[userID]; !ok {
			digestInfos[userID] = &DigestInfo{}
			identityIDs = append(identityIDs, userID)
		}
		digestInfos[userID].boxesInfo = append(digestInfos[userID].boxesInfo, &BoxInfo{ID: boxID})
	}

	return digestInfos, identityIDs, nil

}

// isEligible returns true if the received identity is eligible for current digest frequency
// and remove identity from digestInfo if not eligible
// Eligibility conditions:
// the identity for has a notification configuration which matches the current job
// the identity have not been used by an active user recently
func (dj *DigestJob) isEligible(ctx context.Context, identity identity.Identity) bool {
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
		return false
	}
	return true
}
