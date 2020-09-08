package events

import (
	"context"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type accessContent struct {
	RestrictionType string `json:"restriction_type"`
	Value           string `json:"value"`
}

func addAccessHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, _ entrypoints.IdentityIntraprocessInterface) error {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return merror.Transform(err).Describe("checking admin")
	}

	// check content format
	var c accessContent
	if err := e.JSONContent.Unmarshal(&c); err != nil {
		return merror.Transform(err).Describe("marshalling access content")
	}
	if err := v.ValidateStruct(&c,
		v.Field(&c.RestrictionType, v.Required, v.In("invitation_link", "identifier", "email_domain")),
		v.Field(&c.Value, v.Required),
	); err != nil {
		return merror.Transform(err).Describe("validating access content")
	}

	// check the access doesn't exist yet
	content := e.JSONContent.String()
	_, err := get(ctx, exec, eventFilters{
		eType:     null.StringFrom("access.add"),
		unrefered: true,
		boxID:     null.StringFrom(e.BoxID),
		content:   &content,
	})
	// no error means the access already exists
	if err == nil {
		return merror.Conflict().Describe("this access already exists").
			Detail("content", merror.DVConflict).
			Detail("box_id", merror.DVConflict)
	}
	// a not found is what is expected so we do ignore it
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}
	return nil
}

func rmAccessHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, _ entrypoints.IdentityIntraprocessInterface) error {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return merror.Transform(err).Describe("checking admin")
	}

	if err := v.ValidateStruct(e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return err
	}

	if e.JSONContent.String() != "{}" {
		return merror.BadRequest().Describe("content should be empty").Detail("content", merror.DVForbidden)
	}

	// the referrer must exist and not been refered
	// access.add refered means an access.rm already exist
	_, err := get(ctx, exec, eventFilters{
		eType:     null.StringFrom("access.add"),
		unrefered: true,
		boxID:     null.StringFrom(e.BoxID),
		id:        e.ReferrerID,
	})
	// no error means the access.rm already exists
	if err != nil {
		return merror.Transform(err).Describe("checking access.add referrer_id consistency")
	}
	return nil
}

func FindActiveAccesses(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:     null.StringFrom(boxID),
		eType:     null.StringFrom("access.add"),
		unrefered: true,
	})
}

func MustMemberHaveAccess(
	ctx context.Context,
	exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface,
	boxID string, identityID string,
) error {
	// 1. the identity must be a member of the box
	isMember, err := isMember(ctx, exec, boxID, identityID)
	if err != nil {
		return err
	}
	if !isMember {
		return merror.Forbidden().Describe("must be a member")
	}

	return MustHaveAccess(ctx, exec, identities, boxID, identityID)
}

func MustHaveAccess(
	ctx context.Context,
	exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface,
	boxID string, identityID string,
) error {
	// 1. check some access exists for the box
	accesses, err := FindActiveAccesses(ctx, exec, boxID)
	if err != nil {
		return err
	}

	// 2. if no access exists or the box is closed, only the admins has access to
	closedBox, err := isClosed(ctx, exec, boxID)
	if err != nil {
		return err
	}
	if len(accesses) == 0 || closedBox {
		isAdmin, err := isAdmin(ctx, exec, boxID, identityID)
		if err != nil {
			return err
		}
		if !isAdmin {
			return merror.Forbidden().Describe("must be an admin")
		}
	}

	// 3. consider the box can be public to return directly
	// further security barriers exists because of encryption if the box is public
	// but was not shared
	if isPublic(ctx, accesses) {
		return nil
	}

	// 3. if the box isn't public, get the identity to check whitelist rules
	identity, err := identities.Get(ctx, identityID)
	if err != nil {
		return merror.Transform(err).Describe("getting identity for access check")
	}

	// 4. check restriction rules
	for _, access := range accesses {
		c := accessContent{}
		// on marshal error the box is locked and considered as not public
		if err := access.JSONContent.Unmarshal(&c); err != nil {
			return merror.Transform(err).Describef("access %s corrupted", access.ID)
		}
		switch c.RestrictionType {
		case "identifier":
			if identity.Identifier.Value == c.Value {
				return nil
			}
		case "email_domain":
			if identity.Identifier.Kind == domain.EmailIdentifier &&
				emailHasDomain(identity.Identifier.Value, c.Value) {
				return nil
			}
		}
	}
	return merror.Forbidden().Describe("must match a restriction rule")
}

func isPublic(ctx context.Context, accesses []Event) bool {
	for _, access := range accesses {
		// TODO (perf): save unmarshal into the Event to not re-unmarshal later on
		c := accessContent{}
		// on marshal error the box is locked and considered as not public
		if err := access.JSONContent.Unmarshal(&c); err != nil {
			logger.FromCtx(ctx).Err(err).Msgf("access %s corrupted", access.ID)
			return false
		}
		// if one found restriction is not an invitation_link, the box is not public
		if c.RestrictionType != "invitation_link" {
			return false
		}
	}
	return true
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
