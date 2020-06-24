---
title: SSO - Auth Flow
---

Performing an auth flow is the only to obtain an access token and ID token.
This section provides all the routes the frontend SSO application might use.

All routes described in the sequence diagram are not specified in this document.
Most of them do no reuqire specific frontend logic. Only the user's agent redirect
is required.

Routes are ordered consedering the expected way they should be called.

An auth flow is linked to both authorization and authentication.
A description of these concepts can be found in the [Authorization & Authentication specification](../../concepts/authorization-and-authentication/).
It is probably worth to read before implement following routes.


**Authentication information**:

The final ID Token contains information about the authentication performed by the user.

See [the list of **Authentication Method References** and corresponding **Authentication Context Classes**](http://localhost:1313/concepts/authorization-and-authentication/#43-methods) for more info.

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
        "avatar_url": null,
        "account_id": "5f80b4ec-b42a-4554-a738-4fb532ba2ee4", 
    },
    "authn_step": {
        "identity_id": "4a98b5a1-1c08-46c9-8f26-18d54cbed30a",
        "method_name": "password",
        "metadata": (format described in the next section)
    },
  }
```

- `identity` (object): the authable identity linked to the received identifier value.
  - `display_name` (string): a customizable display name.
  - `avatar_url` (string) (nullable): the web address of the end-user avatar file.
  - `account_id` (string) (nullable): the potential account id linked to the identity.
- `authn_step` (object): the preferred authentication step:
  - `identity_id` (uuid string): the unique identity id the authentication step is attached to.
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_): the preferred authentication method.
  - `metadata` (string) (nullable): filled considering the preferred method.

### Possible formats for the `metadata` field

Considering the preferred authentication method, the metadata can contain additional information.

I - **prehashed_password** as `method_name`:

/!\ Warning, the metadata has not the exact same shape as [the metadata used to perform
an authentication step](./#possible-formats-for-the-metadata-field-1) using the prehased_password method, which also contains the hash of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {{% include "include/hashParameters.json" 2 %}},
  [...]
}
```

## Perform an authentication step in the login flow

The next step to authenticate the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

The metadata field contained in the authentication step depends of the method name.

**Endpoint extensions**:
- Reset the end-user password: while performing a authn step, there is a possibility to reset the password [using a payload extension](./#reset-password-extension).

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
    "metadata": "(format defined in next section)"
  }
}
```

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the performed authentication step information:
  - `identity_id` (uuid string): the authable identity id.
  - `method_name` (string) (one of: _emailed_code_, _prehased_password_): the authentication method used.
  - `metadata` (json object): metadata containing the emailed code value or the prehashed password.
The list of possible formats is defined in the next section.

### Possible formats for the `metadata` field

This section describes the possible metadata format, as a JSON object, which is a
field contained in the JSON body of the previous section.

The context of this specification is the performing of an authentication step only.

I - **emailed_code** as `method_name`:

_JSON Body:_
```json
{
  [...]
  "method_name": "emailed_code",
  "metadata": {
    "code": "320028"
  },
  [...]
}
```

II - **prehashed_password** as `method_name`:

/!\ Warning, the metadata has not the exact same shape as [the metadata returned requiring
an authable identity](./#possible-formats-for-the-metadata-field) with the prehased_password as preferred method, which contains only the hash parameters of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {{% include "include/passwordHash.json" 2 %}},
  [...]
}
```

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

### Reset Password Extension

Reset password is an extension that can be used during the user authentication.

Be sure you have read [the whole route documentation](./#perform-an-authentication-step-in-the-login-flow) to understand it.

Extending the JSON body payload, you can set new value to the account password (and the backup since the backup requires updates when the password changes).

/!\ Today, the extension requires:
- the method `emailed_code` to be used to be accepted by the server.
- the current identity to be linked to an account.

Using this extension, the final token ACR will be considered as if the `prehashed_password` method has been used instead of the `emailed_code` method.

The extension adds to the initial payload explained below
a `password_reset` json object describing the new password and backup values.

The `prehashed_password` contains information following [Argon2 server relief concepts](../../concepts/server-relief/).

_JSON Body:_
```json
  {
    [...] (json payload described in the route below)  
    "password_reset": {
      "prehashed_password": {{% include "include/passwordHash.json" 6 %}},
      "backup_data": "[STRINGIFIED JSON]"
    }
  }
```

- `password_reset` (object): information concerning the new password and backup.
  - `prehashed_password` (object): prehashed password using argon2:
    - `params` (object): argon2 parameters:
      - `memory` (integer).
      - `parallelism` (integer).
      - `iterations` (integer).
      - `salt_base64` (base64 string).
    - `hash_base64` (base64 string): the prehashed password.
  - `backup_data` (string): the new user backup data.


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
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_): the method used by the authentication step.

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
    "desc": "initing an authn step: a code has already been generated",
    "details": {
        "identity_id": "conflict",
        "method_name": "conflict"
    }
}
```

**2. Impossible to perform a prehashed_password method with the identity:**

This error occurs when the authable identity has no linked account. The password being attached to the account, such an authentication method is impossible to be handled.

_Code:_
```bash
  HTTP 409 Conflict
```

```json
{
    "code": "conflict",
    "origin": "body",
    "desc": "initing an authn step: identity has no linked account",
    "details": {
        "identity_id": "conflict",
        "account_id": "required"
    }
}
```

## Accept the consent request in the consent flow

This lets the user choose the scopes they want to accept.

For the moment, those scopes are limited to `tos` and `privacy_policy`

### Request

```bash
  POST https://api.misakey.com.local/auth/consent
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_
```json
{
  "consent_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
  "consented_scopes": [
    "tos",
    "privacy_policy"
  ]
}
```

- `consent_challenge` (string): can be found in previous redirect URL.
- `identity_id` (uuid string): the subject of the flow.
- `consented_scopes` (list of string) (one of: _tos_, _privacy\_policy_): the accepted scopes

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

**1 - A mandatory scope is missing from consent**

At least both `tos` and `privacy_policy` scopes must be consented on this request.

Here is the error to expect if the client didn't send these scopes:
```json
{
     "code": "forbidden",
     "origin": "unknown",
     "details": {
       "tos": "required",
       "privacy_policy": "required"
     },
}
```

Both `tos` and `privacy_policy` are returned, there is no granularity about which one is missing or if it is both of them.

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
