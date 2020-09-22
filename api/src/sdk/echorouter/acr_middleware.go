package echorouter

import (
	"errors"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
)

// NewACRMidlw to handle ACR requirements for given routes
func NewACRMidlw(requiredACR ajwt.ACRSecLvl, strict bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// get accesses, use strict boolean to consider raising an error or not
			claims := ajwt.GetAccesses(c.Request().Context())
			// continue if nil but not strict...
			if claims == nil && !strict {
				return next(c)
			}
			// ...otherwise, raise an error
			if claims == nil {
				c.Error(errors.New("missing accesses from context"))
				return nil
			}
			// check minimal required sec level
			if err := claims.ACRIsGTE(requiredACR); err != nil {
				// service doesn't require any notion of acr today
				if claims.IsNotAnyService() {
					// we raise the ACR error if arrived here
					c.Error(err)
					return nil
				}
			}
			return next(c)
		}
	}
}
