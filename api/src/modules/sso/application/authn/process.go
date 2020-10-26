package authn

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// Process allows to have multi Step in a login flow
// this entity is attached to a login flow and contains information
// about:
// - previous performed Step
// - ACR borne by a potential session
// - expected ACR for the login flow
// - access_token allowing some advanded operation that require authorization
// the Process disappears after the login flow has been accepted
// or after some time
type Process struct {
	LoginChallenge string          `json:"lgc"`
	SessionACR     oidc.ClassRef   `json:"sacr"`
	ExpectedACR    oidc.ClassRef   `json:"eacr"`
	CompleteAMRs   oidc.MethodRefs `json:"camr"`
	IdentityID     string          `json:"mid"`

	AccessToken string `json:"tok"`
	ExpiresAt   int64  `json:"exp"`
	IssuedAt    int64  `json:"iat"`

	// not stored
	NextStep *Step `json:"-"`
}

type processRepo interface {
	Create(context.Context, *Process) error
	Update(context.Context, Process) error
	Get(context.Context, string) (Process, error)
	GetClaims(ctx context.Context, token string) (oidc.AccessClaims, error)
}

func (as *Service) computeNextStep(ctx context.Context, identity identity.Identity, p Process) (Process, error) {
	s, err := as.NextStep(ctx, identity, p.CompleteAMRs.ToACR(), oidc.NewClassRefs(p.ExpectedACR))
	if err != nil {
		return p, merror.Transform(err).Describe("getting next step")
	}
	p.NextStep = &s
	return p, nil
}

// InitProcess and store it
// Set an AMR "BrowserCookie" if sessionACR is not empty.
// NOTE: the identityID is not set by the init of a process today
// in a near future it should be done using the authn session
// today there is no case where the authn session in used in a multi auth step process so there is no need
func (as *Service) InitProcess(ctx context.Context, challenge string, sessionACR, expectedACR oidc.ClassRef) error {
	tok, err := genTok()
	if err != nil {
		return merror.Transform(err).Describe("generating access token")
	}
	p := Process{
		LoginChallenge: challenge,
		SessionACR:     sessionACR,
		ExpectedACR:    expectedACR,

		AccessToken: tok,
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
		IssuedAt:    time.Now().Unix(),
	}
	if sessionACR != oidc.ACR0 {
		p.CompleteAMRs.Add(oidc.AMRBrowserCookie)
	}
	return as.processes.Create(ctx, &p)
}

// UpgradeProcess by adding an amr on it
// it inits the process if required,
// it returns the upgraded Process, telling the login flow require more authn-step to be performed if a NextStep has been set.
func (as *Service) UpgradeProcess(
	ctx context.Context,
	challenge string,
	identity identity.Identity,
	amr oidc.MethodRef,
) (Process, error) {
	process := Process{}
	// retrieve the process
	process, err := as.processes.Get(ctx, challenge)
	if err != nil {
		return process, merror.Transform(err).Describe("getting process")
	}

	// if the process already has an identityID bound to it, check its consistency
	if process.IdentityID != "" && identity.ID != process.IdentityID {
		return process, merror.Forbidden().Describe("cannot change identity id during a process")
	}

	// update the process
	process.CompleteAMRs.Add(amr)
	process.IdentityID = identity.ID
	if err := as.processes.Update(ctx, process); err != nil {
		return process, merror.Transform(err).Describe("updating process")
	}

	// ACR OK
	if process.CompleteAMRs.ToACR() >= process.ExpectedACR {
		return process, nil
	}

	// ACR KO -  compute then return the next authn step
	process, err = as.computeNextStep(ctx, identity, process)
	if err != nil {
		return process, merror.Transform(err).Describe("computing next step")
	}
	return process, nil
}

// genTok returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func genTok() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Note that err == nil only if we read len(b) bytes.
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[0:32], nil
}
