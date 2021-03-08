package org

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

type Org struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatorID string    `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`

	// for now, this is ignored
	Domain  null.String `json:"-"` // https://gitlab.misakey.dev/misakey/user-needs/-/issues/392
	LogoURL null.String `json:"-"` // https://gitlab.misakey.dev/misakey/user-needs/-/issues/395
}

func newOrg() *Org { return &Org{} }

func (o Org) toSQLBoiler() *sqlboiler.Organization {
	return &sqlboiler.Organization{
		ID:        o.ID,
		Name:      o.Name,
		CreatorID: o.CreatorID,
		Domain:    o.Domain,
		LogoURL:   o.LogoURL,
		CreatedAt: o.CreatedAt,
	}
}

func (o *Org) fromSQLBoiler(src sqlboiler.Organization) *Org {
	o.ID = src.ID
	o.Name = src.Name
	o.CreatorID = src.CreatorID
	o.Domain = src.Domain
	o.LogoURL = src.LogoURL
	o.CreatedAt = src.CreatedAt
	return o
}

func Create(ctx context.Context, exec boil.ContextExecutor, creatorID, name string) (Org, error) {
	var o Org

	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return o, merr.From(err).Desc("could not generate uuid v4")
	}

	// init the org
	o = Org{
		ID:        id.String(),
		Name:      name,
		CreatorID: creatorID,
	}

	if err := o.toSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return o, err
	}
	return o, err
}

func GetOrg(ctx context.Context, exec boil.ContextExecutor, id string) (*Org, error) {
	record, err := sqlboiler.FindOrganization(ctx, exec, id)
	if err != nil {
		return nil, err
	}

	return newOrg().fromSQLBoiler(*record), nil
}

func MustBeAdmin(ctx context.Context, exec boil.ContextExecutor, orgID string, identityID string) error {
	org, err := GetOrg(ctx, exec, orgID)
	if err != nil {
		return merr.From(err).Desc("getting organization")
	}
	// identity is admin if it is the creator of the org
	if identityID == org.CreatorID {
		return nil
	}
	// identity is admin if it is an org machine
	if identityID == org.ID {
		return nil
	}
	// otherwise, it is not an admin
	return merr.Forbidden()
}

// TODO (structure): the cache should be refactored into a cross-module package (inside sdk eventually)
func GetIDsForIdentity(ctx context.Context, boxExec boil.ContextExecutor, redConn *redis.Client, identityID string) ([]string, error) {
	pattern := fmt.Sprintf("cache:user_%s:*", identityID)
	keys, err := redConn.Keys(pattern).Result()
	if err != nil {
		return nil, merr.From(err).Desc("listing identity org cache keys")
	}

	orgIDs := []string{}
	// if no keys found, re-build the cache
	if len(keys) == 0 {
		orgMap, err := events.BuildIdentityOrgBoxCache(ctx, boxExec, redConn, identityID)
		if err != nil {
			return nil, merr.From(err).Desc("building identity org cache")
		}
		for orgID := range orgMap {
			orgIDs = append(orgIDs, orgID)
		}
	} else {
		// if cached keys have been found, use it
		for _, key := range keys {
			// cache:user_id:org_{id}:...
			//   0      1      2     x
			splitByColon := strings.Split(key, ":")
			if len(splitByColon) < 3 {
				continue
			}
			orgKey := splitByColon[2]
			// org_{id}...
			//  0    1
			splitByUnderscore := strings.Split(orgKey, "_")
			if len(splitByUnderscore) != 2 {
				continue
			}
			orgIDs = append(orgIDs, splitByUnderscore[1])
		}
	}
	return orgIDs, nil
}

func ListByIDsOrCreatorID(ctx context.Context, exec boil.ContextExecutor, orgIDs []string, creatorID string) ([]Org, error) {
	mods := []qm.QueryMod{}

	if len(orgIDs) > 0 {
		mods = append(mods, sqlboiler.OrganizationWhere.ID.IN(orgIDs))

	}
	// in all cases, use an Or2 to retrieve both by org ids or by creator id
	mods = append(mods, qm.Or2(sqlboiler.OrganizationWhere.CreatorID.EQ(creatorID)))

	records, err := sqlboiler.Organizations(mods...).All(ctx, exec)
	if err != nil {
		return nil, merr.From(err).Desc("querying orgs")
	}

	// build the final org list
	orgs := make([]Org, len(records))
	for idx, record := range records {
		orgs[idx] = *newOrg().fromSQLBoiler(*record)
	}

	return orgs, nil
}
