package events

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
	"gitlab.misakey.dev/misakey/backend/api/src/box/quota"
)

// GetBox ...
func GetBox(ctx context.Context, exec boil.ContextExecutor, identities *IdentityMapper, boxID string, lastEvent *Event) (Box, error) {
	return computeBox(ctx, boxID, exec, identities, lastEvent)
}

// GetBoxWithSenderInfo ...
func GetBoxWithSenderInfo(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, boxID, identityID string) (*Box, error) {
	box, err := computeBox(ctx, boxID, exec, identities, nil)
	if err != nil {
		return nil, err
	}

	// fill the eventCounts attribute
	eventsCount, err := CountEventsBoxForIdentity(ctx, redConn, identityID, boxID)
	if err != nil {
		return nil, merr.From(err).Desc("counting new events")
	}
	box.EventsCount = null.IntFrom(eventsCount)

	boxSetting, err := GetBoxSetting(ctx, exec, identityID, boxID)
	if err != nil {
		return nil, merr.From(err).Desc("getting box setting")
	}
	box.BoxSettings = boxSetting

	return &box, nil
}

// CountBoxesForIdentity returns the number of boxes the identity is concerned by
func CountBoxesForIdentity(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityID, ownerOrgID string, datatagID *string) (int, error) {
	boxIDs, err := listBoxIDsForIdentity(ctx, exec, redConn, identityID, ownerOrgID, datatagID)
	return len(boxIDs), err
}

// ListBoxesForIdentity ...
func ListBoxesForIdentity(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper,
	identityID, ownerOrgID string, datatagID *string,
	limit, offset int,
) ([]*Box, error) {
	boxes := []*Box{}

	// 0. list box ids the identity is concerned by
	allBoxIDs, err := listBoxIDsForIdentity(ctx, exec, redConn, identityID, ownerOrgID, datatagID)
	if err != nil {
		return boxes, err
	}

	// 1. retrieve lastest events concerning the boxes the identity has access to
	list, err := ListLastestForEachBoxID(ctx, exec, allBoxIDs)
	if err != nil {
		return boxes, merr.From(err).Desc("listing box ids")
	}

	// 2. put pagination in place
	// if the offset is higher than the total size, we return an empty list
	if offset >= len(list) {
		return boxes, nil
	}
	// cut the slice using the offset
	list = list[offset:]
	// cut the slice using the limit
	if len(list) > limit {
		list = list[:limit]
	}

	// 3. compute all boxes
	paginatedBoxIDs := make([]string, len(list))
	boxes = make([]*Box, len(list))
	for i, e := range list {
		// TODO (perf): computation in redis
		box, err := computeBox(ctx, e.BoxID, exec, identities, &e)
		if err != nil {
			return boxes, merr.From(err).Descf("computing box %s", e.BoxID)
		}
		boxes[i] = &box
		paginatedBoxIDs[i] = box.ID
	}

	// 4. retrieve box settings
	settingsFilters := BoxSettingFilters{
		BoxIDs:     paginatedBoxIDs,
		IdentityID: identityID,
	}
	boxSettings, err := ListBoxSettings(ctx, exec, settingsFilters)
	if err != nil {
		return boxes, merr.From(err).Desc("listing box settings")
	}
	indexedBoxSettings := make(map[string]BoxSetting, len(boxSettings))
	for _, boxSetting := range boxSettings {
		indexedBoxSettings[boxSetting.BoxID] = *boxSetting
	}

	// 5. add events count and box settings data to boxes
	for _, box := range boxes {
		// we won’t return an error since the list
		// can still be returned
		// with a wrong amount of event counts
		box.EventsCount = null.IntFrom(computeCount(ctx, redConn, identityID, box.ID))

		// add box settings
		boxSetting, ok := indexedBoxSettings[box.ID]
		if !ok {
			boxSetting = *GetDefaultBoxSetting(identityID, box.ID)
		}
		box.BoxSettings = &boxSetting
	}

	return boxes, nil
}

func listBoxIDsForIdentity(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	identityID, ownerOrgID string, datatagID *string,
) ([]string, error) {

	// NOTE: this may be optimized by fetching only the right values from the cache
	// but that would lead to several branches and be less readable
	sortedIDs, err := listBoxIDsForIdentitySortedByOrgAndDatatag(ctx, exec, redConn, identityID)
	if err != nil {
		return []string{}, nil
	}

	// if a datatagID was asked, we return directly the right box IDs
	if datatagID != nil {
		result, ok := sortedIDs[ownerOrgID][*datatagID]
		if !ok {
			return []string{}, nil
		}
		return result, nil
	}

	// else we merge results from all datatags
	var boxIDs []string
	boxIDsforOrg, ok := sortedIDs[ownerOrgID]
	if !ok {
		return []string{}, nil
	}
	for _, ids := range boxIDsforOrg {
		boxIDs = append(boxIDs, ids...)
	}
	return boxIDs, nil
}

func listBoxIDsForIdentitySortedByOrgAndDatatag(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	identityID string,
) (map[string]map[string][]string, error) {
	// 1. try to retrieve and use cache
	cacheBoxIDs := make(map[string]map[string][]string)
	cacheKey := cache.BoxIDsKeysByUser(identityID)
	keys, err := redConn.Keys(cacheKey).Result()
	if err != nil {
		return cacheBoxIDs, nil
	}
	for _, key := range keys {
		keyInfo := strings.Split(key, ":")
		orgID := strings.TrimPrefix(keyInfo[2], "org_")
		datatagID := strings.TrimPrefix(keyInfo[3], "datatag_")
		temp, err := redConn.SMembers(key).Result()
		if err != nil {
			return cacheBoxIDs, nil
		}
		if _, ok := cacheBoxIDs[orgID]; !ok {
			cacheBoxIDs[orgID] = make(map[string][]string)
		}
		cacheBoxIDs[orgID][datatagID] = append(cacheBoxIDs[orgID][datatagID], temp...)
	}
	if len(cacheBoxIDs) != 0 {
		return cacheBoxIDs, nil
	}

	// 2. otherwise, let's build the cache and use its computation
	return BuildIdentityOrgBoxCache(ctx, exec, redConn, identityID)
}

// BuildIdentityOrgBoxCache computes the full cache for the given identity: org_*:boxIDs
// return a map[orgID]boxIDs
func BuildIdentityOrgBoxCache(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	identityID string,
) (map[string]map[string][]string, error) {
	boxIDsByOrgID := make(map[string]map[string][]string)
	// let's build the cache which is organized this way: per user -> per org -> box ids
	// 2. to build the cache means to build the list of user's boxes for all organizations
	// the user's boxes are defined by:
	// - what they have joined (a)
	// - what they have created (b)
	// a.
	activeJoins, err := ListIdentityActiveJoins(ctx, exec, identityID)
	if err != nil {
		return boxIDsByOrgID, merr.From(err).Desc("listing joined box ids")
	}
	// b.
	creates, err := ListCreateByCreatorID(ctx, exec, identityID)
	if err != nil {
		return boxIDsByOrgID, merr.From(err).Desc("listing creator box ids")
	}

	// need to retrieve all create events of the boxes in order to class by org ids
	// 1.a. list the create contents for all the box ids identified for the user
	// it contains org id information that is used to sort the boxes
	boxIDs := make([]string, len(activeJoins)+len(creates))
	idx := 0
	for _, event := range append(activeJoins, creates...) {
		boxIDs[idx] = event.BoxID
		idx++
	}

	// if the identity has access to no box, return directly
	if len(boxIDs) == 0 {
		return boxIDsByOrgID, nil
	}

	contentByBoxID, err := MapCreationContentByBoxID(ctx, exec, boxIDs)
	if err != nil {
		return boxIDsByOrgID, merr.From(err).Desc("listing creation contents")
	}
	// 1.b. sort the boxIDs by orgID and datatagID
	for boxID, createContent := range contentByBoxID {
		if boxIDsByOrgID[createContent.OwnerOrgID] == nil {
			boxIDsByOrgID[createContent.OwnerOrgID] = map[string][]string{}
		}
		// boxes without datatags are stored under the particular "" value
		// because they can be requested
		datatagID := ""
		if createContent.DatatagID != nil {
			datatagID = *createContent.DatatagID
		}
		boxIDsByOrgID[createContent.OwnerOrgID][datatagID] = append(boxIDsByOrgID[createContent.OwnerOrgID][datatagID], boxID)
	}

	// 2. update the cache
	for ownerOrgID, datatag := range boxIDsByOrgID {
		for datatagID, boxIDs := range datatag {
			key := cache.BoxIDsKeyByUserOrgDatatag(identityID, ownerOrgID, datatagID)
			if _, err := redConn.SAdd(key, slice.StringSliceToInterfaceSlice(boxIDs)...).Result(); err != nil {
				logger.FromCtx(ctx).Warn().Err(err).Msgf("could not add boxes cache for identity=%s org=%s", identityID, ownerOrgID)
			}
		}
	}

	// return the cache
	return boxIDsByOrgID, nil
}

// ClearBox ...
func ClearBox(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	// 1. Delete all the events
	if err := DeleteAllForBox(ctx, exec, boxID); err != nil {
		return merr.From(err).Desc("deleting events")
	}

	// 2. Delete the key shares
	if err := keyshares.EmptyAll(ctx, exec, boxID); err != nil {
		return merr.From(err).Desc("emptying keyshares")
	}

	// 3. Delete the box used space
	if err := quota.DeleteBoxUsedSpace(ctx, exec, boxID); err != nil {
		return merr.From(err).Desc("emptying box used space")
	}

	return nil
}

// ListDatatagsForIdentity by getting boxes corresponding to organization
// and extracting all the corresponding datatag
func ListDatatagIDsForIdentity(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityID string, orgID string) ([]string, error) {
	datatagIDs := []string{}
	sortedIDs, err := listBoxIDsForIdentitySortedByOrgAndDatatag(ctx, exec, redConn, identityID)
	if err != nil {
		return datatagIDs, err
	}
	boxIDsforOrg, ok := sortedIDs[orgID]
	if !ok {
		return datatagIDs, nil
	}

	for datatagID := range boxIDsforOrg {
		// we skip the particular datatag "" (no datatag)
		if datatagID != "" {
			datatagIDs = append(datatagIDs, datatagID)
		}
	}

	return datatagIDs, nil
}
