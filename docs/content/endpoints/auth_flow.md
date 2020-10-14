+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Auth Flow endpoints"
tags = ["sso", "authflow", "api", "endpoints"]
title = "SSO - Auth Flow"
+++

## 1. Introduction

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

## 2. Overall auth flow

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
    app.misakey.com->>api.misakey.com: require an authable identity for an identifier
    api.misakey.com->>api.misakey.com: potentially create a new identity/account
    api.misakey.com->>app.misakey.com: returns the authable identity information
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
    api.misakey.com->>app.misakey.com: redirects user's agent to final url with tokens
{{</mermaid>}}

## 3. Initiate an authorization code flow

### 3.1. request

```bash
  GET https://auth.misakey.com/_/oauth2/auth
```

_Query Parameters:_
- see [Open ID Connect RFC](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).

### 3.2. success response

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

## 4. Get Login Information

This route is used to retrieve information about the current login flow using a login challenge.

### 4.1. request

```bash
GET https://api.misakey.com/auth/login/info
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.

### 4.2. response

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

## 5. Require an authable identity for a given identifier

This request is idempotent.

This route is used to retrieve information the authable identity the end-user will log in.

The authable identity can be a new one created for the occasion or an existing one.
See _Response_ below for more information.

### 5.1. request

_Headers:_
- The request doesn't require an authorization header.

```bash
PUT https://api.misakey.com/identities/authable
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

### 5.2. success response

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
    "method_name": "(possibilities described below)",
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
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_, _account_creation): the preferred authentication method.
  - `metadata` (string) (nullable): filled considering the preferred method.

### 5.3. possible formats for the `metadata` field

Considering the preferred authentication method, the metadata can contain additional information.

#### 5.3.1. method name: **emailed_code**

```json
{
  [...]
  "method_name": "emailed_code",
  "metadata": null,
  [...]
}
```

#### 5.3.2. method name: **prehashed_password**

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

#### 5.3.3. method name: **account_creation** :bust_in_silhouette:

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

:information_source: The authorization server might return first an `emailed_code` authn step in order to proove the user owns the identifier.
:mag: The `account_creation` will then come as a second authn step as a [More Authentication Required Response](http://localhost:1313/endpoints/auth_flow/#the-more-authentication-required-response).
:information_source: This step is skipped if the end-user has provided a valid login session corresponding to a previous ACR 1 authentication.

## 6. Perform an authentication step in the login flow

The next step to authenticate the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

The metadata field contained in the authentication step depends of the method name.

### 6.1. request

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
  - `identity_id` (uuid string): the authable identity id.
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_, _account_creation_): the authentication method used.
  - `metadata` (json object): metadata containing the emailed code value or the prehashed password.
The list of possible formats is defined in the next section.

#### 6.1.1. Possible formats for the `metadata` field

This section describes the possible metadata format, as a JSON object, which is a
field contained in the JSON body of the previous section.

The context of this specification is the performing of an authentication step only.

##### 6.1.1.1. method name: **emailed_code**

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

###### 6.1.1.1.1. reset password extension

Reset password is an extension that can be used during the user authentication, when an `emailed_code` authn step is performed.

:warning: The extension requires the current identity to be linked to an account.

Extending the JSON body payload, you can set new value to the account password (and the backup since the backup requires updates when the password changes).
Using this extension, the final token ACR will be considered as if the `prehashed_password` method has been used instead of the `emailed_code` method.

The extension adds to the initial payload explained below
a `password_reset` json object describing the new password and backup values.

:mag: The `prehashed_password` contains information following [Argon2 server relief concepts](../../concepts/server-relief/).

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



##### 6.1.1.2. method name: **prehashed_password**

:warning: Warning, the metadata has not the exact same shape as [the metadata returned requiring
an authable identity](./#possible-formats-for-the-metadata-field) with the `prehashed_password` value as preferred method, which contains only the hash parameters of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {{% include "include/passwordHash.json" 2 %}},
  [...]
}
```

##### 6.1.1.3. method name: **account_creation**


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

### 6.2. success response

On success, the route can return two possible json body:

#### 6.2.1. the "redirect" response

What is returned is the next URL the user's agent should be redirected to.
This response is given when the authentication server consider the end-user has proven its identity sufficiently.

Also part of the response, a CSRF token that must be used for the next request as a `X-CSRF-Token` header. This token prevents from CSRF attacks.

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
  "access_token": "NaEOLikFLklPERz1Cq-umthWAXKBgwWQ-S3cmWqWt8Q.fPAN_Hxyp8eFRQ0zPT5_lNcFANXENcYGlgLiUrS2xY4"
}
```

- `next` (oneof: _redirect_, _authn_step_): the next action the authentication server is waiting for.
- `redirect_to` (string): the URL the user's agent should be redirected to.
- `csrf\_token` (string): The CSRF token that must be sent in a `X-CSRF-Token`

#### 6.2.2. the "more authentication required" response

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

:mag: `method_name` and `metadata` possibilities are defined in [the require authable identity section](#possible-formats-for-the-metadata-field).

### 6.3. notable error responses

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

## 7. Init a new authentication step

This endpoint allows to init an authentication step:
- in case the last one has expired
- if a new step must be initialized

### 7.1. request

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

### 7.2. success response

This route does not return any content.

_Code:_
```bash
HTTP 204 No Content
```

### 7.3. notable error responses

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

## 8. Get Consent Information

This route is used to retrieve information about the current consent flow using a consent challenge.

### 8.1. request

```bash
GET https://api.misakey.com/auth/consent/info
```

_Query Parameters:_
- `consent_challenge` (string): the consent challenge corresponding to the current auth flow.

### 8.2. success response

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

- `subject` (uuid string): unique id of the identifier getting the token
- `acr` (string): the acr level for the current flow
- `scope` (string): list of scope sent during the auth flow init.
- `context` (object): context of the current consent flow
- `client` (object): information about the SSO client involved in the auth flow:
  - `id` (uuid string): the unique id.
  - `name` (string): the name.
  - `url` (string) (nullable): web-address of the logo file.


## 9. Accept the consent request in the consent flow

This lets the user choose the scopes they want to accept.

For the moment, those scopes are limited to `tos` and `privacy_policy`

### 9.1. request

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

### 9.2. success response

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

### 9.3. notable error responses

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

## 10. Logout

This request logouts a user from their authentication session.

An authentication session is valid for an identity but it potentially links other identities
through the account relationship, be aware this action invalidates the session for the whole
account in this case.

### 10.1. request

```bash
POST https://api.misakey.com/auth/logout
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 0): `mid` claim as the identity id sent in body.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

### 10.2. success response

This route does not return any content.

_Code:_
```bash
HTTP 204 No Content
```

## Get Backup

This endpoint allows to get the account backup
during the auth flow.

This endpoint needs a valid process token.

### Request

```bash
GET https://api.misakey.com.local/auth/backup
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.
- `identity_id` (string) (uuid4): the id of the identity corresponding to the current auth flow.

_Headers_:
- `accesstoken` (opaque token) (ACR >= 2): the `login_challenge` token’s claim must be the login challenge sent in query.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

### Success Response

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

## Creating a Backup Key Share

This endpoint allows to create a backup key share in the auth flow.

### Request

```bash
  POST https://api.misakey.com/auth/backup-key-shares
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the access token is given during the auth flow. It must be generated for an identity id linked to the account id sent in the request.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

- `account_id` (string) (uuid): the account for which the shares has been created.
- `share` (string) (base64): one of the shares.
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.

### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```
