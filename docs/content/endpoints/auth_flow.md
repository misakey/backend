---
title: Auth Flow
---

Performing an auth flow is the only to obtain an access token and ID token.
This section provides all the routes the frontend SSO application might use.

All routes described in the sequence diagram are not specified in this document.
Most of them do no reuqire specific frontend logic. Only the user's agent redirect
is required.

Routes are ordered consedering the expected way they should be called.

An auth flow is linked to both authorization and authentication.
A description of these concepts can be found [here](../../concepts/authorization-and-authentication/).
It is probably worth to read before implement following routes.

## Overall auth flow

As of today:
- `auth.misakey.com`: the frontend client
- `auth.misakey.com/_`: the Ory Hydra service
- `api.misakey.com`: the backend service responsible for authentication

{{<mermaid>}}
sequenceDiagram
    auth.misakey.com->>auth.misakey.com/_: initiates oauth2 authorization code
    auth.misakey.com/_->>+auth.misakey.com: redirects the user's agent with login challenge
    Note right of auth.misakey.com: Starts Login Flow
    auth.misakey.com-->>-api.misakey.com: .
    api.misakey.com-->auth.misakey.com/_: fetches login info
    api.misakey.com->>api.misakey.com: checks user's session
    api.misakey.com->>auth.misakey.com: redirect to login page
    auth.misakey.com->>api.misakey.com: require an authable identity for an identifier
    api.misakey.com->>api.misakey.com: potentially create a new identity/account
    api.misakey.com->>auth.misakey.com: returns the authable identity information
    auth.misakey.com->>api.misakey.com: authenticates user with credentials
    api.misakey.com-->auth.misakey.com/_: transmits login info and receives redirect url with login verifier
    api.misakey.com->>+auth.misakey.com: redirects end user to auth server with login verifier
    auth.misakey.com->>-auth.misakey.com/_: .
    Note right of auth.misakey.com: Ends Login Flow
    auth.misakey.com/_->>+auth.misakey.com: redirects the user's agent with consent challenge
    auth.misakey.com->>-api.misakey.com: .
    Note right of auth.misakey.com: Starts Consent Flow
    api.misakey.com-->auth.misakey.com/_: fetches consent info
    api.misakey.com-->auth.misakey.com/_: transmits consent info and receives redirect url with consent verifier
    api.misakey.com->>+auth.misakey.com: redirects end user to auth server with consent verifier
    auth.misakey.com->>-auth.misakey.com/_: .
    Note right of auth.misakey.com: Ends Consent Flow
    api.misakey.com->>+auth.misakey.com: redirects user's agent to redirect url with code
    auth.misakey.com->>-api.misakey.com: .
    api.misakey.com-->auth.misakey.com/_: fetches tokens as an authenticated client
    api.misakey.com->>auth.misakey.com: redirects user's agent to final url with tokens
{{</mermaid>}}

____
# Initiate an authorization code flow

```bash
    GET https://auth.misakey.com.local/_/oauth2/auth
```

  Some query parameters are expected, see [Open ID Connect RFC](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).

### Response

_Code:_
```bash
    HTTP 302 FOUND
```

_Headers:_
```bash
    Location: https://api.misakey.com.local/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e
```

_JSON Body:_
```html
    <a href="https://api.misakey.com.local/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e">Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

____
# Require an authable identity for a given identifier


This request is idempotent.

This route is used to retrieve information about current login flow
and the authable identity the end-user will log in.

The authable identity can be a new one created for the occasion or an existing one.
See _Response_ below for more information.

```bash
  PUT https://api.misakey.com.local/identities/authable
```

```json
  {
  	"login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
  	"identifier": {
  		"value": "auth@test.com"
  	}
  }
```

The request doesn't require an authorization header.

_JSON Body:_

- `login_challenge` (string): can be found in preivous redirect URL.
- `value` (string): the identifier value the end-user entered in the dedicated input text.

### Success Response

This route returns current login flow information and the authable identity the end-user will login as.

_Code:_
```bash
    HTTP 200 OK
```

```json
  {
    "identity": {
        "id": "4a98b5a1-1c08-46c9-8f26-18d54cbed30a",
        "account_id": null,
        "identifer_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
        "is_authable": true,
        "display_name": "auth@test.com",
        "notifications": "minimal",
        "avatar_url": null,
        "confirmed": false
    },
    "login_info": {
      "client_id": "c001d00d-5ecc-beef-ca4e-b00b1e54a111",
      "scope": [
        "openid"
      ],
      "acr_values": null,
      "login_hint": ""
    }
  }
```

- `id` (uuid string): unique identity id.
- `account_id` (uuid string) (nullable): linked account identifier.
- `is_authable` (boolean): either the identity can be used to performed to authenticate the end-user.
- `display_name` (string): a customizable display name.
- `notifications` (string) (one of: _minimal_, _moderate_ or _frequent_): the configuration of notificatons.
- `avatar_url` (string) (nullable): the web address of the end-user avatar file.
- `confirmed` (boolean): either the identity has been proven.
- `client_id` (uuid string): client_id parameter received during the auth flow init.
- `scope` (array of strings) (can be empty): the current auth flow's scope.
- `acr_values` (array of strings) (can be empty): the current auth flow's acr_values.
- `login_hint` (string) (can be empty): the current auth flow's login_hint.

____
# Perform a authentication step in the login flow

The next step to authentication the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

```bash
  POST https://api.misakey.com.local/login/step
```

```json
{
  "login_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "step": {
    "identity_id": "e45f579fd02d41adbf8cb45e0f6a44ff",
    "method_name": "emailed_code",
    "metadata": { "code": "320028" }
  }
}
```

The request doesn't require an authorization header.

_JSON Body:_

- `login_challenge` (string): can be found in preivous redirect URL.
- `identity_id` (uuid string): the authable identity id.
- `method_nam` (string) (one of: _emailed_code_): the authentication method used.
- `metadata` (json object): metadata containing the emailed code value.

### Success Response

On success, the route returns the next URL to redirect the user's agent.

_Code:_
```bash
  HTTP 200 OK
```

_JSON Body:_
```json
{
    "redirect_to": "https://auth.misakey.com.local/_/oauth2/auth"
}
```

- `redirect_to` (string): the URL the user's agent should be redirected to.

### Notable Error Responses

On errors, some information should be displayed to the end-user.

**1. Received code invalid:**

This error occurs when the code received in metadata does not match any stored code.

_Code:_
```bash
  HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
     "code": "forbidden",
     "origin": "body",
     "details": {
       "code": "invalid",
     },
}
```

**2. Received code expired:**

This error occurs when the code received in metadata is correct but the timebox
to use it is expired.

_Code:_
```bash
  HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
     "code": "forbidden",
     "origin": "body",
     "details": {
       "code": "expired",
     },
}
```
