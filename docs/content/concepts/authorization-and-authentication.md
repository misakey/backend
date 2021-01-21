+++
categories = ["Concepts"]
date = "2020-09-11"
description = "Authorization & Authentication"
tags = ["concepts", "authorization", "authentication"]
title = "Authorization & Authentication"
+++

## 1. Introduction

_Disclaimer: this document is intended for internal Misakey people and curious fellows
to understand how our auth system works.
It has not been built as a documentation to help third-parties integrate the Misakey auth system._

## 2. Lexicon

- **Authentication/AuthN**: process to verify the identity of a caller.
- **Authorization/AuthZ**: process to authorize a consumer (frontend clients, backend services...)
to access a resource.
- **Client**: a third-party, a relying party.
- **User**: a human being (the one called: internet citizen).

## 3. Authorization

### 3.1. Basics

Requests are authorized using bearer tokens and following [OAuth 2.0][] protocol,
by default the system uses the [OpenID Connect][] protocol to allow authentication
verification by our consumers.

The system uses [Ory Hydra][], an open source Go implementation of these protocols,
certified by the [Open ID foundation][].

To generate a token for a user, **only the authorization code flow using a
confidential client** is possible.

The exchange of the code against the final tokens needs [client authentication][],
the `private_key_jwt` method only can be used.

### 3.2. Token

There are two types of tokens obtained at the end of an authorization code flow:
- an `access Token`, which allows any client to obtain authorizations to Misakey
  resources on behalf of the end user.
  **This access token is now a `HttpOnly` cookie and as a result it is managed entirely by the browser (`HttpOnly` actually makes it inaccessible for the JavaScript)**
- an `identity Token (ID Token)`, described in [the authentication][] section,
  which allows identification of the authenticated user and contains many information
  such as when the token expires or how secure was the authentication process
  (using or not 2FA, as an example).

The access token is opaque, bearing authorizations to access resources.
It cannot be introspected from the external world.

Tokens (ID & access tokens) lifetime is one hour.

### 3.3. Scopes

Regarding [OAuth 2.0][] protocol, the system relies on the [scope parameter][] parameter
to generate tokens linked to more advanced authorization concepts.

`openid` is default and required. It enables the [OpenID Connect][] layer
on top of [OAuth 2.0][] to generates an [ID Token][] aside the access token.

`tos` and `privacy_policy` are legal scopes that ensure
the user has consented to Term of Services and Privacy Policy.

### 3.4. TOS and Privacy Policy

`tos` and `privacy_policy` scopes can be requested during the init of the authorization code flow using the parameter `scope`.

In the Misakey app, these scopes are managed for a given account. It means that if one identity of an account
has an existing valid session with those scopes, these will be automatically accepted
for all identities of the account.

For all other External Relying Parties, these scopes are managed for a given identity.

If the scopes have been requested but have never been accepted, the user will be redirected to a frontend page
during the consent flow and the route `POST /auth/consent` must be used to
publish their consent.

If requested, the legal scopes must be consented by the end-user during the consent flow.

## 4. Authentication

### 4.1. Token

There are two types of tokens obtained at the end of an authorization code flow:
- an `access Token`, which is an authorization token (see [the authorization][] section).
- an `identity Token (ID Token)`, which allows the identification of the authenticated user
  and contains many information such as when the token expires
  or how secure was the authentication process.

The identity token is a signed [JWT][] containing information
about who is authenticated and the authentication process used.

Specification of the [ID Token][] content (all claims are described in the RFC).


**Claims working with Misakey's business logic:**

- The system uses the [RFC Authentication Method Reference Values][] to set `amr` claim
- described in the [ID Token][] or [the methods][] section.

- The system uses the `acr` claim (a.k.a. Authentication Context Class)
of the ID token to indicate to the third-party how secure was the authentication process
to guarantee the caller identity. It is the Security Level (or Sec Level).
So **ACR = Sec Level**. `acr` value is a number from 1 to 10.
1 being the less secure authentication flow, 10 being a certainty
about the user identity who just authenticated.

:warning: Only `sco` is an additional claim containing scopes linked to the access token,
it can contain default scopes, caller types, consented purposes or assigned roles.

ID Token example:
```json
{
  "acr": "4",
  "amr": [
    "pwd",
    "webauthn"
  ],
  "sco": "openid",
  "at_hash": "z-N-JLaW6RbtqWfqTABIrw",
  "aud": [
    "00000000-0000-0000-0000-000000000000"
  ],
  "auth_time": 1568883569,
  "exp": 1568887169,
  "iat": 1568883569,
  "iss": "https://auth.misakey.com/_/",
  "jti": "8a45c7d5-dbe1-41ea-8d89-f8d4359958b9",
  "nonce": "",
  "rat": 1568883569,
  "sid": "a90615fc-8ce1-4f19-ae40-75c2228b199e",
  "sub": "18b88d48-ad48-43b9-a323-eab1de68b280"
}
```

:warning: If the involved SSO Client is Misakey, there are additional claims in the ID Token:
```json
{
  "mid": "33345c7d5-dbe1-41ea-8d89-f8d4359958b9",
  "aid": "218234bb-7cd5-4211-a25d-3ee3337cd03fd"
}
```

- `mid` (string) (uuid): the identity id bound to the token (mid stands for Misakey ID since it is the main ID in Misakey).
- `aid` (string) (uuid) (nullable): the account id bound to the token.

Tokens (ID & access tokens) lifetime is one hour.

### 4.2. Authentication Request

A certain level of acr/security can be asked while initiating an authorization code flow,
using the `acr_values` described in [the authentication request][] section.

:warning: To specify multiple values (by order of preference, says the RFC) will result
in only taking into account the first one, we consider third-parties forces
the security level and it is not up to the end-user to choose.

:warning: Third-parties are responsible of managing potential session with lower `acr` than requested.
If an authentication session is still active, `acr_values` parameter might be ignored
since only the ID token is aware of the previous `acr` value and so the Misakey auth server is not.
Final `acr` claims must be always compared to initial `acr_values`.
If the final `acr` shows a less secure authentication than expected:
the authorization code flow must be performed again using `prompt=login` parameter
(or force authentication and ignore potential session.

### 4.3. Methods

#### 4.3.1. Introduction

Authentication methods, also named A.M.R. for Authentication Method Reference,
are different way/path to prove a user identity.

The system follows, as an example, [Level of Assurance][] paper to consider
how an authentication method is secure.

**What kind of authentication method should be used ?**

The kind of authentication method for a given `acr` is today known by both auth server & the user interfaces.

The auth server currently check for the authentication method to correspond
to the requested `acr_values`.

Table of correspondance between ACR and Authentication methods:

- `browser_cookie`: depends of the authentication method used to generate the session.
- `emailed_code`: ACR 1.
- `prehashed_password`: ACR 2.
- `prehashed_password` + `totp`: ACR 3.
- `prehashed_password` + `webauthn`: ACR 4.

#### 4.3.2. Browser Cookie

Browser Cookie allows the client to not re-ask to the end-user to login, it
also make possible silent authentication to get a new access token smoothly.

Long-lived browser cookies aren't a decent authentication method to ensure
the connected user is the owner of the identity, still the corresponding `acr`
if set accoridng to the authentication method used to generate the session.

#### 4.3.3. Emailed Code

Emailed Code method is a randomly generated 6 digits code sent to the user
using a external channel.

To enforce it, during the init of the auth flow:
- `acr_values` query parameter must be set to `1`.
- `prompt` query parameter must be set to `login`.

To perform it:
- The auth step method name must be: `emailed_code`.

The user must complete the authentication by sending the received code in a short time window.
The user must wait some timee between two codes generation/sending.

Used alone, its final corresponding `acr` is 1.

#### 4.3.4. Password

Password method is a hashed password comparison, using [Argon2 server relief][].

To enforce it, during the init of the auth flow:
- `acr_values` query parameter must be set to `2`.
- `prompt` query parameter must be set to `login`.

To perform it:
- The auth step method name must be: `prehashed_password`.

Used alone, its final corresponding `acr` is 2.

#### 4.3.4. (WIP) TOTP

[TOTP][] is a way to achieve multi-factor authentication using a specific algorithm generating time-based one time password.

To enforce it, during the init of the auth flow:
- `acr_values` query parameter must be set to `3`.
- `prompt` query parameter must be set to `login`.
- the end-user must always have an account configured with `totp`. `webauthn` method might be used instead (cf dedicated section) and will result with an acr 4 (with a stronger authentication then).
:warning: If the user hasn't configure any mfa method, this method cannot be performed and the final acr will be set according to the previous entered method (1, 2...).

To perform it:
- The auth step method name must be: `totp`.

It is always combined with a `prehashed_password` first. Its final corresponding `acr` is 3.

#### 4.3.4. (WIP) Webauthn

[Webauthn][] is a way to achieve multi-factor authentication using public-key cryptography.

To enforce it, during the init of the auth flow:
- `acr_values` query parameter must be set to `4` or `3`.
- `prompt` query parameter must be set to `login`.
- the end-user must always have an account configured with `webauthn` method.
- :warning: If the user hasn't configure any mfa method, this method cannot be performed and the final acr will be set according to the previous entered method (1, 2...).

To perform it:
- The auth step method name must be: `webauthn`.

It is always combined with a `prehashed_password` first. Its final corresponding `acr` is 4.

### 4.4 ACR Errors Handling

Aside obvious errors such as "invalid secrets" or "email not existing within our system"
that may occur during the auth flow, some errors specific to ACR should be known
and user interfaces should consider them:

During any call to any endpoint requiring an authorization bearer token,
an error might be raised if the token ACR is not high enough to access the resource:
```
{
  "code": "forbidden",
  "origin": "acr",
  "desc": "token acr is too weak",
  "details": {
    "acr": "forbidden",
    "required_acr": "2"
  }
}
```

Good to know:
- `origin` is set to `acr` to allow specific error handling on this kind of error.
It should be checked first by consumer (user interface).
- `required_acr` tells the consumer (user interface) what `acr_values` should be
while initing the auth flow to access the resource.

[OAuth 2.0]: https://tools.ietf.org/html/rfc6749
[OpenID Connect]: https://openid.net/specs/openid-connect-core-1_0.html
[Ory Hydra]: https://www.ory.sh/docs/hydra
[Open ID foundation]: https://openid.net/certification
[client authentication]: https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication
[the authentication]: #4-authentication
[the authorization]: #3-authorization
[JWT]: https://tools.ietf.org/html/rfc7519
[scope parameter]: https://tools.ietf.org/html/rfc6749#section-3.3
[ID Token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[RFC Authentication  Method Reference Values]: https://tools.ietf.org/html/rfc8176
[the methods]: (#431-introduction)
[the authentication request]: https://openid.net/specs/openid-connect-core-1_0.html#ImplicitAuthRequest
[Level of Assurance]: https://www.itu.int/rec/T-REC-X.1254-201209-I/en
[Argon2 server relief]: https://password-hashing.net/submissions/specs/Argon-v3.pdf
[TOTP]: https://tools.ietf.org/html/rfc6238
[Webauthn]: https://webauthn.guide/

## 5. CSRF Protection

The access token is stored in well-protected cookies (`http-only`, `secure` and `SameSite=strict` cookies), but some browser does not support those secure options.
The cookies are thus still exposed to CSRF attacks.

That is why every authenticated call must be protected by an anti-CSRF token (we will call it a *CSRF Token* for the sake of simplicity).

The CSRF Token is sent at the end of the auth flow and must be sent with each request in the header `X-CSRF-Token`.

It must be stored somewhere safe in the app.
