package authz

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type tokenIntrospector interface {
	GetBearerTok(ctx echo.Context) (value string, fromCookie bool, err error)
	GetClaims(ctx context.Context, token string) (oidc.AccessClaims, error)
	CheckClientID(ctx context.Context, claims oidc.AccessClaims) error
	HandleErr(eCtx echo.Context, next echo.HandlerFunc, err error) error
}

// NewTokenIntrospector is a middleware used to declare than a route require authorization.
// The opaque token is found, instropected and information are set inside the current request context
// to be checked later by different actors (modules...)
// the way of retrieval, checks... of bearer tokens are defined by the given token repo
func NewTokenIntrospector(
	mode, selfCliID string, selfCliOnly bool,
	tokenRepo interface{}, redConn *redis.Client,
) echo.MiddlewareFunc {
	// init the introspector considering the mode
	var manager tokenIntrospector
	var err error
	switch mode {
	case "hydra":
		manager, err = newHydraIntrospector(tokenRepo, selfCliID, selfCliOnly)
	case "authn_process":
		manager, err = newAuthnProcessInstropector(tokenRepo, selfCliID)
	default:
		err = merr.Internal().Descf("invalid mode %s for token introspector", mode)
	}
	if err != nil {
		log.Fatalf("encountered error while creating the introspector: %s", err.Error())
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(eCtx echo.Context) error {
			// handle bearer token
			opaqueTok, fromCookie, err := manager.GetBearerTok(eCtx)
			if err != nil {
				return manager.HandleErr(eCtx, next, err)
			}

			ctx := eCtx.Request().Context()

			// introspect the token and get the claims
			acc, err := manager.GetClaims(ctx, opaqueTok)
			if err != nil {
				if merr.IsAnInternal(err) { // internal is directly returned - not special manager handling
					return merr.From(err).Desc("introspecting token")
				}
				return manager.HandleErr(eCtx, next, err)
			}

			// check the client id is authorized to perform this request
			if err := manager.CheckClientID(ctx, acc); err != nil {
				fmt.Println("error on client id check:", err)
				return manager.HandleErr(eCtx, next, err)
			}

			// set access claims in request context
			eCtx.SetRequest(eCtx.Request().WithContext(oidc.SetAccesses(ctx, &acc)))

			// if the requester is a machine, return directly
			if IsAMachine(acc) {
				return next(eCtx)
			}

			// if the requester is not an machine
			// ensure the token has been sent using cookie (csrf is used by a upper middleware in this case)
			// this is mandatory
			if !fromCookie {
				return merr.Unauthorized().Desc("access token must be sent using cookies and a csrf token")
			}

			//store last interaction data - ignore error because this information can be lost
			_ = redConn.Set(fmt.Sprintf("lastInteraction:user_%s", acc.IdentityID), time.Now().Unix(), 0*time.Second)

			return next(eCtx)
		}
	}
}
