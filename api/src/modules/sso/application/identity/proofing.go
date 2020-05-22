package identity

import (
	"context"
	"time"

	"github.com/volatiletech/null"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type identityProofRepo interface {
	Create(*domain.IdentityProof) error
	Update(context.Context, *domain.IdentityProof) error
	List(context.Context, domain.IdentityProofFilters) ([]*domain.IdentityProof, error)
}

func (ids *IdentityService) InitEmailCodeProofing(identity domain.Identity) error {
	codeMetadata, err := generateCodeMetadata()
	if err != nil {
		return err
	}

	flow := domain.IdentityProof{
		IdentityID: identity.ID,
		MethodName: domain.EmailCodeMethod,
		Metadata:   codeMetadata,

		InitiatedAt: time.Now(),

		Asserted:   false,
		AssertedAt: null.Time{},
	}
	if err := ids.proofs.Create(&flow); err != nil {
		return err
	}
	return nil
}

// Assert the identity proof considering the method name and the received metadata
// Return true in case of success, otherwise returns false with a correspond error
// Updates the identity proof entity, but do not confirm the identity itself
func (ids *IdentityService) Assert(ctx context.Context, assertion domain.IdentityProof) (bool, error) {
	// get last initiated proofs orbered by the most recent
	existings, err := ids.proofs.List(ctx, domain.IdentityProofFilters{
		IdentityID: &assertion.IdentityID,
		MethodName: &assertion.MethodName,
		LastFirst:  null.BoolFrom(true),
	})
	if err != nil {
		return false, err
	}
	if len(existings) == 0 {
		return false, merror.NotFound().Describe("corresponding proof not found").
			Detail("id", merror.DVNotFound).Detail("method_name", merror.DVNotFound)
	}
	// take the most recent proof - ignore others
	mostRecentProof := existings[0]
	// check the most recent proof has not been already asserted
	if mostRecentProof.Asserted {
		return false, merror.Conflict().Describe("most recent proof already asserted")
	}

	// check the metadata
	var metadataErr error
	switch mostRecentProof.MethodName {
	case domain.EmailCodeMethod:
		// transform metadata into code metadata structure
		input, err := toCodeMetadata(assertion.Metadata)
		if err != nil {
			metadataErr = merror.Forbidden().From(merror.OriBody).
				Describe(err.Error()).Detail("metadata", merror.DVMalformed)
			break
		}
		stored, err := toCodeMetadata(mostRecentProof.Metadata)
		if err != nil {
			metadataErr = merror.Forbidden().Describe(err.Error()).Detail("stored_code", merror.DVMalformed)
			break
		}

		// compare codes
		if input.Code != stored.Code {
			metadataErr = merror.Forbidden().From(merror.OriBody).Detail("metadata", merror.DVInvalid)
			break
		}

		// assert the identity proof and update the entity
		mostRecentProof.Asserted = true
		mostRecentProof.AssertedAt = null.TimeFrom(time.Now())
		if err := ids.proofs.Update(ctx, mostRecentProof); err != nil {
			return false, err
		}
		return true, nil
	default:
		metadataErr = merror.BadRequest().Detail("method_name", merror.DVInvalid)
	}
	return false, metadataErr
}
