package authn

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
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
	PasswordReset  bool            `json:"pwdr"`
	AccountID      string          `json:"aid"`

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

// InitProcess and store it
// Set an AMR "BrowserCookie" if sessionACR is not empty.
// NOTE: the identityID is not set by the init of a process today
// in a near future it should be done using the authn session
// today there is no case where the authn session in used in a multi auth step process so there is no need
func (as *Service) InitProcess(
	ctx context.Context,
	challenge string, sessionACR, expectedACR oidc.ClassRef,
) error {
	tok, err := genTok()
	if err != nil {
		return merr.From(err).Desc("generating access token")
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

// GetProcess using the login challenge
func (as *Service) GetProcess(
	ctx context.Context,
	challenge string,
) (Process, error) {
	// retrieve the process
	process, err := as.processes.Get(ctx, challenge)
	if err != nil {
		return process, merr.From(err).Desc("getting process")
	}
	return process, nil
}

// UpdateProcess to change its state manually
func (as *Service) UpdateProcess(
	ctx context.Context, redConn *redis.Client,
	challenge string, expectedACR oidc.ClassRef, passwordReset bool,
) error {
	process := Process{}
	// retrieve the process
	process, err := as.processes.Get(ctx, challenge)
	if err != nil {
		return merr.From(err).Desc("getting process")
	}

	process.PasswordReset = passwordReset
	process.ExpectedACR = expectedACR
	if err := as.processes.Update(ctx, process); err != nil {
		return merr.From(err).Desc("updating process")
	}

	return nil
}

// UpgradeProcess by adding an amr on it
// it inits the process if required,
// it returns the upgraded Process, telling the login flow require more authn-step to be performed if a NextStep has been set.
func (as *Service) UpgradeProcess(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	challenge string, identity identity.Identity, amr oidc.MethodRef,
) (Process, error) {
	// retrieve the process
	process, err := as.processes.Get(ctx, challenge)
	if err != nil {
		return process, merr.From(err).Desc("getting process")
	}

	// if the process already has an identityID bound to it, check its consistency
	if process.IdentityID != "" && identity.ID != process.IdentityID {
		return process, merr.Forbidden().Desc("cannot change identity id during a process")
	}

	// if the process already has an account ID bound to it, check its consistency
	if process.AccountID != "" && identity.AccountID.String != process.AccountID {
		return process, merr.Forbidden().Desc("cannot change account id during a process")
	}

	// update the process
	process.CompleteAMRs.Add(amr)
	process.IdentityID = identity.ID
	process.AccountID = identity.AccountID.String
	if err := as.processes.Update(ctx, process); err != nil {
		return process, merr.From(err).Desc("updating process")
	}

	// get potential next step - can be nil
	process.NextStep, err = as.PrepareNextStep(
		ctx, exec, redConn,
		identity, process.CompleteAMRs.ToACR(), process.ExpectedACR,
		process.PasswordReset,
	)
	if err != nil {
		return process, merr.From(err).Desc("getting next step")
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
