package application

import (
	"context"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type IdentityProofAssertion struct {
	domain.IdentityProof
}

func (sso SSOService) ConfirmIdentity(ctx context.Context, proofAssertion domain.IdentityProof) error {
	// assert the proof
	asserted, err := sso.identityService.Assert(ctx, proofAssertion)
	if err != nil {
		return err
	}
	if asserted != true {
		return merror.Internal().Describe("assertion not possible - no error detected")
	}

	// once the proof is valid, we can confirm the identity with an update
	return sso.identityService.Confirm(ctx, proofAssertion.IdentityID)
}
