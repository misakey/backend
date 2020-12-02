package events

import (
	"context"
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

type accessContent struct {
	RestrictionType string `json:"restriction_type"`
	Value           string `json:"value"`
	AutoInvite      bool   `json:"auto_invite"`
}

func doAddAccess(ctx context.Context, e *Event, extraJSON null.JSON, exec boil.ContextExecutor, redConn *redis.Client, identityMapper *IdentityMapper, cryptoActionRepo external.CryptoActionRepo, _ files.FileStorageRepo) (Metadata, error) {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	// check content format
	var c accessContent
	if err := e.JSONContent.Unmarshal(&c); err != nil {
		return nil, merror.Transform(err).Describe("unmarshalling access content")
	}
	if err := v.ValidateStruct(&c,
		v.Field(&c.RestrictionType, v.Required, v.In("invitation_link", "identifier", "email_domain")),
		v.Field(&c.Value, v.Required.When(c.RestrictionType != "invitation_link")),
	); err != nil {
		return nil, merror.Transform(err).Describe("validating access content")
	}

	// check the access doesn't exist yet
	content := e.JSONContent.String()
	if c.RestrictionType == "invitation_link" {
		content = fmt.Sprintf("{\"restriction_type\":\"%s\"}", c.RestrictionType)
	}

	_, err := get(ctx, exec, eventFilters{
		eType:      null.StringFrom("access.add"),
		unreferred: true,
		boxID:      null.StringFrom(e.BoxID),
		content:    &content,
	})
	// no error means the access already exists
	if err == nil {
		return nil, merror.Conflict().Describe("this access already exists").
			Detail("content", merror.DVConflict).
			Detail("box_id", merror.DVConflict)
	}
	// a not found is what is expected so we do ignore it
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return nil, err
	}

	// assuming exec is a transaction
	// and caller will roll it back if we return an error
	err = e.persist(ctx, exec)
	if err != nil {
		return nil, err
	}

	if c.RestrictionType == "identifier" {
		// potential side effects of an "identifier" access
		// (auto invitation)
		if c.AutoInvite {
			if extraJSON.Valid {
				box, err := Compute(ctx, e.BoxID, exec, identityMapper, nil)
				if err != nil {
					return nil, merror.Transform(err).Describe("computing the box (to get title for notif)")
				}
				// creates a crypto action AND the notification
				err = cryptoActionRepo.CreateInvitationActions(ctx, e.SenderID, e.BoxID, box.Title, c.Value, extraJSON)
				if err != nil {
					return nil, merror.Transform(err).Describe("creating crypto actions")
				}
			} else {
				return nil, merror.BadRequest().Detail("extra", merror.DVRequired)
			}
		} else if extraJSON.Valid {
			return nil, merror.BadRequest().Detail("auto_invite", merror.DVInvalid)
		}
	}

	return nil, nil
}

func doRmAccess(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, _ *IdentityMapper, _ external.CryptoActionRepo, _ files.FileStorageRepo) (Metadata, error) {
	// the user must be an admin
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking admin")
	}

	if err := v.ValidateStruct(e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return nil, err
	}

	if e.JSONContent.String() != "{}" {
		return nil, merror.BadRequest().Describe("content should be empty").Detail("content", merror.DVForbidden)
	}

	// the referrer must exist and not been referred yet or it is already removed
	// access.add referred means an access.rm already exist for it
	_, err := get(ctx, exec, eventFilters{
		eType:      null.StringFrom("access.add"),
		unreferred: true,
		boxID:      null.StringFrom(e.BoxID),
		id:         e.ReferrerID,
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("checking access.add referrer_id consistency")
	}

	return nil, e.persist(ctx, exec)
}

func FindActiveAccesses(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:      null.StringFrom(boxID),
		eType:      null.StringFrom("access.add"),
		unreferred: true,
	})
}

func MustMemberHaveAccess(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper,
	boxID string, identityID string,
) error {
	// 1. the identity must have access to the box
	if err := MustHaveAccess(ctx, exec, identities, boxID, identityID); err != nil {
		return err
	}

	// 2. the identity must be a member of the box
	isMember, err := isMember(ctx, exec, redConn, boxID, identityID)
	if err != nil {
		return err
	}
	if !isMember {
		return merror.Forbidden().Describe("must be a member").Detail("reason", "not_member")
	}

	return nil
}

func MustHaveAccess(
	ctx context.Context,
	exec boil.ContextExecutor, identities *IdentityMapper,
	boxID string, identityID string,
) error {
	// 1. admin is always allowed to see the box
	IsAdmin, err := IsAdmin(ctx, exec, boxID, identityID)
	if err != nil {
		return err
	}
	if IsAdmin {
		return nil
	}

	// 2. check some access exists for the box
	accesses, err := FindActiveAccesses(ctx, exec, boxID)
	if err != nil {
		return err
	}

	// 3. if no access exists, only the admins has access to
	if len(accesses) == 0 {
		return merror.Forbidden().Describe("must be an admin").Detail("reason", "no_access")
	}

	// 4. if the box is closed, only the admins has access to
	closedBox, err := isClosed(ctx, exec, boxID)
	if err != nil {
		return err
	}

	if closedBox {
		return merror.Forbidden().Describe("cannot access a closed box").Detail("reason", "closed")
	}

	// 5. consider the box can be public to return directly
	// further security barriers exists because of encryption if the box is public
	// but was not shared
	if isPublic(ctx, accesses) {
		return nil
	}

	// 6. if the box isn't public, get the identity to check whitelist rules
	identity, err := identities.Get(ctx, identityID, true)
	if err != nil {
		return merror.Transform(err).Describe("getting identity for access check")
	}

	// 7. check restriction rules
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
			if identity.Identifier.Kind == "email" &&
				emailHasDomain(identity.Identifier.Value, c.Value) {
				return nil
			}
		}
	}
	return merror.Forbidden().Describe("must match a restriction rule").Detail("reason", "no_access")
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
