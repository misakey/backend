package oidc

import (
	"context"
	"testing"
	"time"

	customclock "github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

func TestValid(t *testing.T) {
	tests := map[string]struct {
		claims       AccessClaims
		nowTimestamp int64
		err          error
		// now    time.Time
	}{
		"token is valid": {
			claims: AccessClaims{
				Subject:    "this_is_the_sub",
				IdentityID: "this_is_the_identity",
			},
			nowTimestamp: 0,
		},
		"token is expired": {
			claims: AccessClaims{
				ExpiresAt: 1,
			},
			nowTimestamp: 2,
			err:          merr.Unauthorized().Descf("token expired"),
		},
		"token is used before issued time": {
			claims: AccessClaims{
				ExpiresAt: 3,
				IssuedAt:  2,
			},
			nowTimestamp: 1,
			err:          merr.Unauthorized().Descf("token used before issued"),
		},
		"token is not valid yet": {
			claims: AccessClaims{
				ExpiresAt: 3,
				IssuedAt:  0,
				NotBefore: 2,
			},
			nowTimestamp: 1,
			err:          merr.Unauthorized().Descf("token not valid yet"),
		},
		"token has en empty sub": {
			claims: AccessClaims{
				Subject: "",
			},
			nowTimestamp: 0,
			err:          merr.Unauthorized().Descf("empty subject"),
		},
		"token has en empty identity id": {
			claims: AccessClaims{
				Subject: "this_is_the_sub",
			},
			nowTimestamp: 0,
			err:          merr.Unauthorized().Descf("empty mid"),
		},
	}
	clockMock := customclock.NewMock()
	clock = clockMock
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			// set mocked time
			clockMock.Set(time.Unix(test.nowTimestamp, 0))

			// call the function to test
			err := test.claims.Valid()

			// check expectations
			assert.Equal(t, test.err, err)
		})
	}
}

func TestSetGetAccesses(t *testing.T) {
	t.Run("should retrieve claims for user 1", func(t *testing.T) {
		ctx := SetAccesses(context.Background(), &AccessClaims{Subject: "1"})
		result := GetAccesses(ctx)
		assert.Equalf(t, &AccessClaims{Subject: "1"}, result, "did not get right accessclaims")
	})
}

func TestValidAudience(t *testing.T) {
	tests := map[string]struct {
		claims      AccessClaims
		expectedAud string
		err         error
	}{
		"audience is valid": {
			claims: AccessClaims{
				Audiences: []string{"audience_1", "audience_2"},
			},
			expectedAud: "audience_1",
			err:         nil,
		},
		"audience in not valid": {
			claims: AccessClaims{
				Audiences: []string{"audience_1", "audience_2"},
			},
			expectedAud: "audience_3",
			err:         merr.Unauthorized().Descf("client is not part of the audience"),
		},
		"audience in valid - expectedAud being empty": {
			claims: AccessClaims{
				Audiences: []string{"audience_1", "audience_2"},
			},
			expectedAud: "",
			err:         nil,
		},
	}
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			result := test.claims.ValidAudience(test.expectedAud)
			assert.Equal(t, test.err, result)
		})
	}
}
