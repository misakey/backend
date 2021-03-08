package events

import (
	"context"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

type accessAddContent struct {
	RestrictionType string `json:"restriction_type"`
	Value           string `json:"value"`
	AutoInvite      bool   `json:"auto_invite"`
}

const (
	restrictionIdentifier  = "identifier"
	restrictionEmailDomain = "email_domain"
)

// Unmarshal a access.add content JSON into its typed structure
func (c *accessAddContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

// Validate a access.add content structure
func (c accessAddContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.RestrictionType, v.Required, v.In(restrictionIdentifier, restrictionEmailDomain)),
		v.Field(&c.Value, v.Required),
	)
}

func doAddAccess(
	ctx context.Context, e *Event, extraJSON null.JSON,
	exec boil.ContextExecutor, redConn *redis.Client, identityMapper *IdentityMapper, cryptoRepo external.CryptoRepo, _ files.FileStorageRepo,
) (Metadata, error) {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merr.From(err).Desc("checking admin")
	}

	// check content format
	var access accessAddContent
	if err := e.JSONContent.Unmarshal(&access); err != nil {
		return nil, merr.From(err).Desc("unmarshalling access content")
	}

	// check the access.add doesn't exist yet
	_, err := get(ctx, exec, eventFilters{
		eType:           null.StringFrom(etype.Accessadd),
		boxID:           null.StringFrom(e.BoxID),
		restrictionType: null.StringFrom(access.RestrictionType),
		accessValue:     null.StringFrom(access.Value),
		// exclude removed access.add events for that box
		excludeOnRef: &referentsFilters{
			eTypes: []string{etype.Accessrm},
			boxID:  null.StringFrom(e.BoxID),
		},
	})
	// no error means the access already exists
	if err == nil {
		return nil, merr.Conflict().Desc("this access already exists").
			Add("content", merr.DVConflict).
			Add("box_id", merr.DVConflict)
	}
	// a not found is what is expected so we do ignore it
	if err != nil && !merr.IsANotFound(err) {
		return nil, err
	}

	// assuming exec is a transaction
	// and caller will roll it back if we return an error
	err = e.persist(ctx, exec)
	if err != nil {
		return nil, err
	}

	if access.RestrictionType == "identifier" {
		// potential side effects of an "identifier" access
		// (auto invitation)
		if access.AutoInvite {
			if extraJSON.Valid {
				box, err := computeBox(ctx, e.BoxID, exec, identityMapper, nil)
				if err != nil {
					return nil, merr.From(err).Desc("computing the box (to get title for notif)")
				}
				guest, err := identityMapper.GetByIdentifierValue(ctx, access.Value)
				if err != nil {
					return nil, merr.From(err).Desc("getting guest by identifier value")
				}
				// creates a crypto action AND the notification
				err = CreateInvitationActions(ctx, cryptoRepo, identityMapper, box, guest, e.SenderID, extraJSON, false)
				if err != nil {
					return nil, merr.From(err).Desc("creating crypto actions")
				}
			} else {
				return nil, merr.BadRequest().Add("extra", merr.DVRequired)
			}
		} else if extraJSON.Valid {
			return nil, merr.BadRequest().Add("auto_invite", merr.DVInvalid)
		}
	}

	return nil, nil
}

func doRmAccess(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merr.From(err).Desc("checking admin")
	}

	if err := v.ValidateStruct(e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return nil, err
	}

	if e.JSONContent.String() != "{}" {
		return nil, merr.BadRequest().Desc("content should be empty").Add("content", merr.DVForbidden)
	}

	// the referred access.add event must exist and...
	_, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Accessadd),
		boxID: null.StringFrom(e.BoxID),
		id:    e.ReferrerID,
		// exclude removed access.add events for that box
		excludeOnRef: &referentsFilters{
			eTypes: []string{etype.Accessrm},
			boxID:  null.StringFrom(e.BoxID),
		},
	})
	// access.add referred means an access.rm already exists for it
	if err != nil {
		return nil, merr.From(err).Descf("checking %s referrer_id consistency", etype.Accessadd)
	}

	return nil, e.persist(ctx, exec)
}

// FindActiveAccesses ...
func FindActiveAccesses(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Accessadd),
		// exclude removed access.add events for that box
		excludeOnRef: &referentsFilters{
			eTypes: []string{etype.Accessrm},
			boxID:  null.StringFrom(boxID),
		},
	})
}

// MustBoxExists ...
func MustBoxExists(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	_, err := get(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Create),
	})
	if err != nil {
		return merr.From(err).Desc("getting box create event")
	}
	return nil
}

// CanJoin returns no error if the received identityID can joined the box related to the received box id
func MustBeAbleToJoin(
	ctx context.Context,
	exec boil.ContextExecutor, identities *IdentityMapper,
	boxID, identityID string,
) error {
	// 1. check if the box is in public mode
	if isPublic(ctx, exec, boxID) {
		return nil
	}

	// 2. from here, the box is considered a in limited mode
	// check the identity id has access because of any rule
	return HasAccess(ctx, exec, identities, boxID, identityID, false)
}

// IsLegimate returns no error if the received identityID is legitimate to be in the box
func MustBeLegitimate(
	ctx context.Context,
	exec boil.ContextExecutor, identities *IdentityMapper,
	boxID, identityID string,
) error {
	// 1.. check if the identity id is an admin
	isAdmin, err := IsAdmin(ctx, exec, boxID, identityID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	// 1. check the identity id has access because of an identifier restriction rule
	return HasAccess(ctx, exec, identities, boxID, identityID, true)
}

// hasAccess returns no error if the received identityID match an active access rule
// identifierOnly set to true restricts the matching of access rules to only identifier restriction type.
func HasAccess(ctx context.Context,
	exec boil.ContextExecutor, identities *IdentityMapper,
	boxID, identityID string,
	identifierOnly bool,
) error {
	// 1. check if data subjet
	hasAccess, err := isDataSubject(ctx, exec, boxID, identityID)
	if err != nil {
		return err
	}
	if hasAccess {
		return nil
	}

	// 2. list existing active accesses for the box
	accesses, err := FindActiveAccesses(ctx, exec, boxID)
	if err != nil {
		return err
	}

	// 3. if no access exists, no-one can join it
	if len(accesses) == 0 {
		return merr.Forbidden().Desc("must be an admin").Add("reason", "no_access")
	}

	// 4. get the identity to check whitelist rules
	identity, err := identities.Get(ctx, identityID, true)
	if err != nil {
		return merr.From(err).Desc("getting identity for access check")
	}

	// 5. check restriction rules
	for _, access := range accesses {
		c := accessAddContent{}
		// on marshal error the box is locked and considered as not joinable
		if err := access.JSONContent.Unmarshal(&c); err != nil {
			return merr.From(err).Descf("access %s corrupted", access.ID)
		}
		switch c.RestrictionType {
		case restrictionIdentifier:
			if identity.IdentifierValue == c.Value {
				return nil
			}
		case restrictionEmailDomain:
			// ignore this restriction type if only identifier restriction is requested to be checked
			if identifierOnly {
				continue
			}
			if identity.IdentifierKind == "email" &&
				emailHasDomain(identity.IdentifierValue, c.Value) {
				return nil
			}
		}
	}
	return merr.Forbidden().Desc("must match a restriction rule").Add("reason", "no_access")
}

func isPublic(ctx context.Context, exec boil.ContextExecutor, boxID string) bool {
	// get() always get last event corresponding to the query
	accessModeEvent, err := get(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Stateaccessmode),
	})
	// NOTE: no access mode event found means the default mode is enabled: limited.
	if merr.IsANotFound(err) {
		return false
	}

	c := AccessModeContent{}
	// on marshal error the box is locked and considered as not public
	if err := accessModeEvent.JSONContent.Unmarshal(&c); err != nil {
		logger.FromCtx(ctx).Err(err).Msgf("access mode %s corrupted", accessModeEvent.ID)
		return false
	}
	if c.Value == PublicMode {
		return true
	}
	return false
}

func emailHasDomain(email, domain string) bool {
	domainIndex := strings.LastIndex(email, "@")
	// LastIndex returns -1 on not found @ (invalid email in this case)
	if domainIndex == -1 {
		return false
	}
	// compare stricly domains
	return email[domainIndex+1:] == domain
}
