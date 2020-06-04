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
- `app.misakey.com`: the frontend client
- `auth.misakey.com/_`: the Ory Hydra service
- `api.misakey.com`: the backend service responsible for authentication

{{<mermaid>}}
sequenceDiagram
    app.misakey.com->>auth.misakey.com/_: initiates oauth2 authorization code
    auth.misakey.com/_->>+app.misakey.com: redirects the user's agent with login challenge
    Note right of app.misakey.com: Starts Login Flow
    app.misakey.com-->>-api.misakey.com: .
    api.misakey.com-->auth.misakey.com/_: fetches login info
    api.misakey.com->>api.misakey.com: checks user's session
    api.misakey.com->>app.misakey.com: redirect to login page
    app.misakey.com->>api.misakey.com: require an authable identity for an identifier
    api.misakey.com->>api.misakey.com: potentially create a new identity/account
    api.misakey.com->>app.misakey.com: returns the authable identity information
    app.misakey.com->>api.misakey.com: authenticates user with credentials
    api.misakey.com-->auth.misakey.com/_: transmits login info and receives redirect url with login verifier
    api.misakey.com->>+app.misakey.com: redirects end user to auth server with login verifier
    app.misakey.com->>-auth.misakey.com/_: .
    Note right of app.misakey.com: Ends Login Flow
    auth.misakey.com/_->>+app.misakey.com: redirects the user's agent with consent challenge
    app.misakey.com->>-api.misakey.com: .
    Note right of app.misakey.com: Starts Consent Flow
    api.misakey.com-->auth.misakey.com/_: fetches consent info
    api.misakey.com-->auth.misakey.com/_: transmits consent info and receives redirect url with consent verifier
    api.misakey.com->>+app.misakey.com: redirects end user to auth server with consent verifier
    app.misakey.com->>-auth.misakey.com/_: .
    Note right of app.misakey.com: Ends Consent Flow
    api.misakey.com->>+app.misakey.com: redirects user's agent to redirect url with code
    app.misakey.com->>-api.misakey.com: .
    api.misakey.com-->auth.misakey.com/_: fetches tokens as an authenticated client
    api.misakey.com->>app.misakey.com: redirects user's agent to final url with tokens
{{</mermaid>}}

## Initiate an authorization code flow

### Request

```bash
    GET https://auth.misakey.com.local/_/oauth2/auth
```

_Query Parameters:_
- see [Open ID Connect RFC](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).

### Response

_Code:_
```bash
    HTTP 302 FOUND
```

_Headers:_
- `Location`: https://api.misakey.com.local/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e

_HTML Body:_
```html
    <a href="https://api.misakey.com.local/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e">Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

## Get Login Information

This route is used to retrieve information about the current login flow using a login challenge.

### Request

```bash
  GET https://api.misakey.com.local/auth/login/info
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.

### Response

_Code_:
```bash
    HTTP 200 OK
```

_JSON Body_:
```json
  {
      "client": {
          "id": "c001d00d-5ecc-beef-ca4e-b00b1e54a111",
          "name": "Misakey App",
          "logo_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png"
      },
      "scope": [
          "openid"
      ],
      "acr_values": null,
      "login_hint": ""
  }
```

- `client` (object): information about the SSO client involved in the auth flow:
  - `id` (uuid string): the unique id.
  - `name` (string): the name.
  - `url` (string) (nullable): web-address of the logo file.
- `scope` (string): list of scope sent during the auth flow init.
- `acr_values` (string) (nullable): list of acr values sent during the auth flow init.
- `login_hint` (string): the login_hint sent during the auth flow init.

## Require an authable identity for a given identifier

This request is idempotent.

This route is used to retrieve information the authable identity the end-user will log in.

The authable identity can be a new one created for the occasion or an existing one.
See _Response_ below for more information.

### Request

_Headers:_
- The request doesn't require an authorization header.

```bash
  PUT https://api.misakey.com.local/identities/authable
```

_JSON Body:_
```json
  {
  	"login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
  	"identifier": {
  		"value": "auth@test.com"
  	}
  }
```

- `login_challenge` (string): can be found in preivous redirect URL.
- `identifier` (object): information about the used identifier to authenticate the end-user:
  - `value` (string): the identifier value the end-user entered in the dedicated input text.

### Success Response

This route returns the authable identity the end-user will login as.

_Code:_
```bash
    HTTP 200 OK
```

_JSON Body:_
```json
  {
    "identity": {
        "display_name": "MUCHMICHMACH@test.com",
        "avatar_url": null
    },
    "authn_step": {
        "identity_id": "4a98b5a1-1c08-46c9-8f26-18d54cbed30a",
        "method_name": "emailed_code"
    },
  }
```

- `identity` (object): the authable identity linked to the received identifier value.
  - `display_name` (string): a customizable display name.
  - `avatar_url` (string) (nullable): the web address of the end-user avatar file.
- `authn_step` (object): the preferred authentication step:
  - `identity_id` (uuid string): the unique identity id the authentication step is attached to.
  - `method_name` (string) (one of: _emailed_code_): the authentication method.

## Perform an authentication step in the login flow

The next step to authenticate the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

### Request

```bash
  POST https://api.misakey.com.local/auth/login/authn-step
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_
```json
{
  "login_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "authn_step": {
    "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
    "method_name": "emailed_code",
    "metadata": { "code": "320028" }
  }
}
```

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the performed authentication step information:
  - `identity_id` (uuid string): the authable identity id.
  - `method_name` (string) (one of: _emailed_code_): the authentication method used.
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

## Init a new authentication step

This request allows to init an authentication step:
- in case the last one has expired
- if a new step must be initialized

```bash
  POST https://api.misakey.com.local/authn-steps
```

```json
  {
    "login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
    "authn_step": {
      "identity_id": "fed6784f-913b-49cb-9174-a8b7dc6bc675",
      "method_name": "emailed_code"
    }
  }
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the initiated authentication step information:
  - `identity_id` (uuid string): the identity ID for which the authentication step will be initialized.
  - `method_name` (string): the method used by the authentication step.

### Success Response

This route does not return any content.

_Code:_
```bash
    HTTP 204 No Content
```

### Notable Error Responses

On errors, some information should be displayed to the end-user.

**1. A authn step already exists:**

This error occurs when an authentication step already exists for this `identity_id` and `method_name`

_Code:_
```bash
  HTTP 409 Conflict
```

```json
{
    "code": "conflict",
    "origin": "body",
    "desc": "could not ask for a code: a code has already been generated",
    "details": {
        "identity_id": "conflict",
        "method_name": "conflict"
    }
}
```

## Logout

This request logouts a user from their authentication session.

An authentication session is valid for an identity but it potentially links other identities
through the account relationship, be aware this action invalidates the session for the whole
account in this case.

### Request

```bash
  POST https://api.misakey.com.local/logout
```

_Headers_:
- `Authorization` (opaque token) (ACR >= 0): `subject` claim as the identity id sent in body.

### Success Response

This route does not return any content.

_Code:_
```bash
    HTTP 204 No Content
```
