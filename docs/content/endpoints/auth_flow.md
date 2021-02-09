+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Auth Flow endpoints"
tags = ["sso", "authflow", "api", "endpoints"]
title = "SSO - Auth Flow"
+++

# 1. Introduction

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

See [the list of **Authentication Method References** and corresponding **Authentication Context Classes**](../../concepts/authorization-and-authentication/#43-methods) for more info.

## 1.1. Overall auth flow

As of today:
- `app.misakey.com`: the frontend client
- `auth.misakey.com/_`: the Ory Hydra service
- `api.misakey.com`: the backend service responsible for authentication

{{<mermaid>}}
sequenceDiagram
    app.misakey.com->>auth.misakey.com/_: initiates oauth2 authorization code
    auth.misakey.com/_->>+app.misakey.com: redirects the user's agent with login challenge
    Note right of app.misakey.com: Starts Login Flow
    app.misakey.com-->>-api.misakey.com: 
    api.misakey.com-->auth.misakey.com/_: fetches login info
    api.misakey.com->>api.misakey.com: checks user login sessions
    api.misakey.com->>app.misakey.com: redirect to login page
    app.misakey.com->>api.misakey.com: require an identity for an identifier
    api.misakey.com->>api.misakey.com: potentially create a new identity/account
    api.misakey.com->>app.misakey.com: returns the identity information
    app.misakey.com->>api.misakey.com: authenticates user with credentials
    api.misakey.com-->auth.misakey.com/_: transmits login info and receives redirect url with login verifier
    api.misakey.com->>+app.misakey.com: redirects end user to auth server with login verifier
    app.misakey.com->>-auth.misakey.com/_: 
    Note right of app.misakey.com: Ends Login Flow
    auth.misakey.com/_->>+app.misakey.com: redirects the user's agent with consent challenge
    app.misakey.com->>-api.misakey.com: 
    Note right of app.misakey.com: Starts Consent Flow
    api.misakey.com-->auth.misakey.com/_: fetches consent info
    api.misakey.com-->api.misakey.com: check user consent sessions
    api.misakey.com->>app.misakey.com: redirect to consent page
    app.misakey.com->>api.misakey.com: consent to some scopes
    api.misakey.com-->auth.misakey.com/_: transmits consent info and receives redirect url with consent verifier
    api.misakey.com->>+app.misakey.com: redirects end user to auth server with consent verifier
    app.misakey.com->>-auth.misakey.com/_: 
    Note right of app.misakey.com: Ends Consent Flow
    api.misakey.com->>+app.misakey.com: redirects user's agent to redirect url with code
    app.misakey.com->>-api.misakey.com: 
    api.misakey.com-->auth.misakey.com/_: fetches tokens as an authenticated client
    api.misakey.com->>app.misakey.com: redirects user's agent to final url with ID Token (and access token as cookie)
{{</mermaid>}}

## 1.2. Initiate an authorization code flow

### 1.2.1. request

```bash
  GET https://auth.misakey.com/_/oauth2/auth
```

_Query Parameters:_
- see [Open ID Connect RFC](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).

### 1.2.2. success response

_Code:_
```bash
HTTP 302 FOUND
```

_Headers:_
- `Location`: https://api.misakey.com/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e

_HTML Body:_
```html
    <a href="https://api.misakey.com/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e">Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

# 2. Login Flow

## 2.1. Get Login Information

This route is used to retrieve information about the current login flow using a login challenge.

### 2.1.1. request

```bash
GET https://api.misakey.com/auth/login/info
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.

### 2.1.2. response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "client": {
    "id": "00000000-0000-0000-0000-000000000000",
    "name": "Misakey App",
    "logo_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png",
    "tos_uri": "https://about.misakey.com/#/fr/legals/tos/",
    "policy_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png"
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
  - `logo_uri` (string) (nullable): web-address of the logo file.
  - `policy_uri` (string) (nullable): web-address of client privacy policy.
  - `tos_uri` (string) (nullable): web-address of the client TOS.
- `scope` (string): list of scope sent during the auth flow init.
- `acr_values` (string) (nullable): list of acr values sent during the auth flow init.
- `login_hint` (string): the login_hint sent during the auth flow init.

## 2.2. Require an identity for a given identifier

This request is idempotent.

This route is used to retrieve information the identity the end-user will log in.

The identity can be a new one created for the occasion or an existing one.
See _Response_ below for more information.

### 2.2.1. request

_Headers:_
- The request doesn't require an authorization header.

```bash
PUT https://api.misakey.com/auth/identities
```

_JSON Body:_
```json
{
	"login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
  "identifier_value": "auth@test.com",
  "password_reset": false
}
```

- `login_challenge` (string): can be found in preivous redirect URL.
- `identifier_value` (string): the identifier value the end-user has entered.
- `password_reset` (bool): a boolean to initiate a new password reset flow.

### 2.2.2. success response

This route returns the identity the end-user will login as.

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
    "method_name": "(possibilities described below)",
    "metadata": (format described in the next section)
  },
}
```

- `identity` (object): the identity linked to the received identifier value.
  - `display_name` (string): a customizable display name.
  - `avatar_url` (string) (nullable): the web address of the end-user avatar file.
  - `account_id` (string) (nullable): the potential account id linked to the identity.
- `authn_step` (object): the preferred authentication step:
  - `identity_id` (uuid string): the unique identity id the authentication step is attached to.
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_, _account_creation): the preferred authentication method.
  - `metadata` (string) (nullable): filled considering the preferred method.

### 2.2.3. possible formats for the `metadata` field

Considering the preferred authentication method, the metadata can contain additional information.

#### 2.2.3.1. method name: **emailed_code**

```json
{
  [...]
  "method_name": "emailed_code",
  "metadata": null,
  [...]
}
```

#### 2.2.3.2. method name: **prehashed_password**

On `prehashed_password`, the `metadata` field contains information about how the password is supposed to be prehashed.

:warning: Warning, the metadata has not the exact same shape as [the metadata used to perform
an authentication step](./#possible-formats-for-the-metadata-field-1) using the `prehashed_password` method, which also contains the hash of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {{% include "include/hashParameters.json" 2 %}},
  [...]
}
```

#### 2.2.3.3. method name: **account_creation** :bust_in_silhouette:

```json
{
  [...]
  "method_name": "account_creation",
  "metadata": null,
  [...]
}
```

This is the only way to create an account linked to an identity.

:information_source: This method is returned by the server in a certain configuration.

This method is retured when:
- the final ACR is expected to be 2.
- the identity linked to the given identifier has no linked account.
- the end-user has provided a valid login session corresponding to a previous ACR 1 authentication.

:information_source: The authorization server might return first an `emailed_code` authn step in order to prove the user owns the identifier.
:mag: The `account_creation` will then come as a second authn step as a [More Authentication Required Response](http://localhost:1313/endpoints/auth_flow/#622-the-more-authentication-required-response).
:information_source: This step is skipped if the end-user has provided a valid login session corresponding to a previous ACR 1 authentication.

#### 2.2.3.4. method name: **webauthn**


```json
{
  [...]
  "method_name": "webauthn",
  "metadata":
      {
          "publicKey": {
              "challenge":"<string>",
              "timeout":<int>,
              "rpId":"<string>",
              "allowCredentials":[{
                  "type":"<string>",
                  "id":"<string>"
              },{
                  "type":"<string>",
                  "id":"<string>"
              }]
          }
    },
    [...]
}
```

The metadata content is defined and explained in the webauthn documentation.

#### 2.2.3.5. method name: **totp**


```json
{
  [...]
  "method_name": "totp",
  "metadata": null,
    [...]
}
```

#### 2.2.3.6. method name: **reset_password**

This method must only be used at the end of the flow.

This will set a new password for the given identity.

```json
{

  [...]
  "method_name": "reset_password",
  "metadata": null,
    [...]
}
```

## 2.3. Perform an authentication step in the login flow

The next step to authenticate the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

The metadata field contained in the authentication step depends of the method name.

### 2.3.1. request

```bash
POST https://api.misakey.com/auth/login/authn-step
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_
```json
{
  "login_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "authn_step": {
    "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
    "method_name": "(possibilities defined in the next section)",
    "metadata": "(formats defined in the next section)"
  }
}
```

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the performed authentication step information:
  - `identity_id` (uuid string): the identity id.
  - `method_name` (string) (one of: _emailed\_code_, _prehashed\_password_, _account\_creation_, _webauthn_, _totp_): the authentication method used.
  - `metadata` (json object): metadata containing the emailed code value, the prehashed password or the webauthn options.
The list of possible formats is defined in the next section.

#### 2.3.1.1. Possible formats for the `metadata` field

This section describes the possible metadata format, as a JSON object, which is a
field contained in the JSON body of the previous section.

The context of this specification is the performing of an authentication step only.

##### 2.3.1.1.1. method name: **emailed_code**

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

##### 2.3.1.1.2. method name: **prehashed_password**

:warning: Warning, the metadata has not the exact same shape as [the metadata returned requiring
an identity](./#possible-formats-for-the-metadata-field) with the `prehashed_password` value as preferred method, which contains only the hash parameters of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {{% include "include/passwordHash.json" 2 %}},
  [...]
}
```

##### 2.3.1.1.3. method name: **account_creation**


```json
{
  [...]
  "method_name": "account_creation",
  "metadata": {
    "prehashed_password": {{% include "include/passwordHash.json" 4 %}},
    "backup_data": "[STRINGIFIED JSON]"
    },
  [...]
}
```

##### 2.3.1.1.4. method name: **webauthn**


```json
{
  [...]
  "method_name": "webauthn",
  "metadata": {
      "id":"<string>",
      "rawId":"<string>",
      "response":{
          "clientDataJSON":"<string>",
          "authenticatorData":"<string>",
          "signature":"<string>",
          "userHandle":"<string>"
      },
      "type":"<string>"
    },
  [...]
}
```

The metadata content is explained in the webauthn documentation

##### 2.3.1.1.5. method name: **totp**


```json
{
  [...]
  "method_name": "totp",
  "metadata": {
      "code": "<string>",
      "recovery_code": "<string>"
    },
  [...]
}
```

The `code` must be the otp given by the external app.

The `recovery_code` must be in the user recovery codes set.

One of `code` or `recovery_code` is required, but the other one must be blank.

##### 2.3.1.1.6. method name: **reset_password**

_JSON Body:_
```json
  {
    "method_name": "reset_password",
    "metadata": {
      "prehashed_password": {{% include "include/passwordHash.json" 6 %}},
      "backup_data": "[STRINGIFIED JSON]"
    }
  [...]
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

:mag: The `prehashed_password` contains information following [Argon2 server relief concepts](../../concepts/server-relief/).

### 2.3.2. success response

On success, the route can return two possible json body:

#### 2.3.2.1. the "redirect" response

What is returned is the next URL the user's agent should be redirected to.
This response is given when the authentication server consider the end-user has proven its identity sufficiently.

The access token is sent and stored in an http-only cookie.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "next": "redirect",
  "redirect_to": "https://auth.misakey.com/_/oauth2/auth",
  "csrf_token": "3eb193b251a24dcb8bb5aa5c3cca3487"
}
```

- `next` (oneof: _redirect_, _authn_step_): the next action the authentication server is waiting for.
- `redirect_to` (string): the URL the user's agent should be redirected to.

#### 2.3.2.2. the "more authentication required" response

What is returned is the next authentication step the end-user should perform.
This response is given when the authentication server requires more authentication step to proove the end-user identity.

Technically, it happens when the current ACR, corresponding to previously validated authentication steps (a.k.a. AMRs), is below the expected ACR.

The expected ACR can be set by:
- the Relying Party expectations: `acr_values` paramenter on the init of the auth flow.
- the choosen identity configuration.

Also part of the response, an access token that must be used for the next request as an `accesstoken` cookie. This token allows us to authorize more advanced calls.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "next": "authn_step",
  "authn_step": {
    "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
    "method_name": "...",
    "metadata": "..."
  },
  "access_token": "NaEOLikFLklPERz1Cq-umthWAXKBgwWQ-S3cmWqWt8Q.fPAN_Hxyp8eFRQ0zPT5_lNcFANXENcYGlgLiUrS2xY4"
}
```

- `next` (oneof: _redirect_, _authn_step_): the next action the authentication server is waiting for.
- `access_token` (string): an access token allowing more advanced requests while being still in the login flow. Should be used as `Authorization` header.
- `authn_step` (object): the next expected authn step to end the login flow.

:mag: `method_name` and `metadata` possibilities are defined in [the require identity section](#possible-formats-for-the-metadata-field).

### 2.3.3. notable error responses

On error during an authentication step, some information might be displayed to the end-user.

**1. Received code is invalid:**

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

**2. Received code has expired:**

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

**3. The Authorization headers do not correspond to the login_challenge:**

Situation when the error is returned:
1. The end-user has performed an authentication step in a login flow A.
2. The end-user refreshes their agent, it inits a new login flow B.
3. The client has still the access token bound to the login flow A in headers.

The access token is valid for the login flow A but cannot be used in the login flow B.

_Code:_
```bash
HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
  "code": "forbidden",
  "origin": "headers",
  "desc": "...",
  "details": {
    "Authorization": "conflict",
    "login_challenge": "conflict"
  }
}
```

## 2.4. Init a new authentication step

This endpoint allows to init an authentication step:
- in case the last one has expired
- if a new step must be initialized

### 2.4.1. request

```bash
POST https://api.misakey.com/authn-steps
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

### 2.4.2. success response

This route does not return any content.

_Code:_
```bash
HTTP 204 NO CONTENT
```

### 2.4.3. notable error responses

On errors, some information should be displayed to the end-user.

**1. A similar authn step already exists:**

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

This error occurs when the identity has no linked account. The password being attached to the account, such an authentication method is impossible to be handled.

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

# 3. Consent Flow

## 3.1. Get Consent Information

This route is used to retrieve information about the current consent flow using a consent challenge.

### 3.1.1. request

```bash
GET https://api.misakey.com/auth/consent/info
```

_Query Parameters:_
- `consent_challenge` (string): the consent challenge corresponding to the current auth flow.

### 3.1.2. success response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "subject": "f058a198-e5e7-4d96-8b71-e6aa1edd3eb5",
  "acr": "2",
  "scope": [
    "openid",
    "tos",
    "privacy_policy"
  ],
  "context": {
    "amr": "emailed_code"
  },
  "client": {
    "id": "00000000-0000-0000-0000-000000000000",
    "name": "Misakey App",
    "logo_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png"
  }
}
```

- `subject` (uuid string): unique id of the account getting the token
- `acr` (string): the acr level for the current flow
- `scope` (string): list of scope sent during the auth flow init.
- `context` (object): context of the current consent flow
- `client` (object): information about the SSO client involved in the auth flow:
  - `id` (uuid string): the unique id.
  - `name` (string): the name.
  - `url` (string) (nullable): web-address of the logo file.


## 3.2. Accept the consent request in the consent flow

This lets the user choose the scopes they want to accept.

For the moment, those scopes are limited to `tos` and `privacy_policy`

### 3.2.1. request

```bash
POST https://api.misakey.com/auth/consent
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
- `identity_id` (uuid string): the identity id bound to the identifier of the flow.
- `consented_scopes` (list of string) (one of: _tos_, _privacy\_policy_): the accepted scopes.

### 3.2.2. success response

On success, the route returns the next URL to redirect the user's agent.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "redirect_to": "https://auth.misakey.com/_/oauth2/auth"
}
```

- `redirect_to` (string): the URL the user's agent should be redirected to.

### 3.2.3. notable error responses

**1 - A mandatory scope is missing from consent**

If some legal scopes have been requested at the init of the auth flow, they
must be consented in all cases on this request.

Here is the error to expect if the client didn't send these scopes:
```json
{
  "code": "forbidden",
  "origin": "unknown",
  "details": {
    "requested_legal_scope": "{space-limited list of legal scope requested}",
    "consented_legal_scope": "{space-limited list of legal scope consented}"
  }
}
```

# 4. Others

## 4.1 Reset the auth flow

This requests allow the complete restart of the auth flow. It triggers a redirection to the initial
auth request if found (using the `login_challenge` sent in parameter).
If no auth request is found, it redirects the end-user to a blank connection screen on the Misakey main app (without any information about the initial flow).

:warning: be aware this action invalidates the session for the whole
account in this case.

### 1.2.1. request

```bash
GET https://api.misakey.com/auth/reset?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e
```

_Query Parameters:_
- `login_challenge` (string) (optional): the login challenge corresponding to the current auth flow. On invalid or missing, the user's agent will be redirected to the home page.

### 1.2.2. success response

_Code:_
```bash
HTTP 302 FOUND
```

_Headers:_
- `Location`: https://auth.misakey.com/_/oauth2/auth

_HTML Body:_
```html
    <a href="https://auth.misakey.com/_/oauth2/auth/>Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

## 4.1. Logout

This request logouts a user from their authentication session.

An authentication session is valid for an identity but it potentially links other identities
through the account relationship, be aware this action invalidates the session for the whole
account in this case.

### 4.1.1. request

```bash
POST https://api.misakey.com/auth/logout
```


- `accesstoken` (opaque token) (ACR >= 0): `mid` claim as the identity id sent in body.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

### 4.1.2. success response

This route does not return any content.

_Code:_
```bash
HTTP 204 NO CONTENT
```

## 4.2. Get Backup

This endpoint allows to get the account backup
during the auth flow.

This endpoint needs a valid process token.

### 4.2.1. Request

```bash
GET https://api.misakey.com.local/auth/backup
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.
- `identity_id` (string) (uuid4): the id of the identity corresponding to the current auth flow.

_Headers_:
- `Authorization`: should be `Bearer {opaque_token}` with opaque token being the `login_challenge` of the auth flow.

### 4.2.2. Success Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "data": "[STRINGIFIED JSON]",
    "version": 3,
    "account_id": "d8aa7d0f-81fe-4e66-99d5-fe2b31360ae0"
}
```

- `data` (string): the user backup data.
- `version` (integer): the current backup version.
- `account_id` (string) (uuid4): the id of the account owning the backup.

## 4.3. Creating a Backup Key Share

This endpoint allows to create a backup key share in the auth flow.

### 4.3.1. Request

```bash
  POST https://api.misakey.com/auth/backup-key-shares
```

_Headers_:
- `Authorization`: should be `Bearer {opaque_token}` with opaque token being the access token given during the auth flow. the backup-key-shares must be generated for an identity id linked to the account id bound to the token.

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

- `account_id` (string) (uuid): the account for which the shares has been created.
- `share` (string) (base64): one of the shares.
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.

### 4.3.2. Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

# 5. OIDC endpoints

These endpoints are openid RFC-compliant endpoints.

## 5.1. Get User Info

This endpoint basically allow to get some of the ID token info.

It must be authenticated.

### 5.1.1. request

```bash
GET https://api.misakey.com/auth/userinfo
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1)
- `tokentype`: must be `bearer`

### 5.1.2. response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "acr": "2",
  "aid": "97db9036-4190-4374-b0d1-0775b55f4e94",
  "amr": [
    "browser_cookie"
  ],
  "email": "joni@misakey.com",
  "mid": "aba11ab7-a077-4520-b76c-dba9fac01693",
  "sco": "openid tos privacy_policy",
  "sid": "562e4b78-86d2-4674-b80a-79dcfffb5d38",
  "sub": "2828b8d0-439c-4326-b49b-2736bd6eacb7"
}
```

- `acr` (string): ACR corresponding to the currrent token,
- `amr` (string array): the way the user got the token
- `email` (string): the user email,
- `sco` (string): list of space separated scopes linked to the token,
- `sid` (string): the session identifier,
- `sub` (string): the user identifier,
- `mid` (string): the Misakey specific identifier,
- `aid` (string): the user account identifier,
