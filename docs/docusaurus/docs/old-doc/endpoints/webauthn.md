---
title: Webauthn Configuration
---

## Introduction

### Concept

Webauthn is used in our **2FA** mechanism.

It must register credentials linked to a specific device before being able to login with it.

These endpoints allows to manipulate those credentials.

## Webauthn Credentials

#### Request new Webauthn credentials creation

This initiates a Webauthn registration flow in order to attach webauthn credentials to the identity.

#### Request

```bash
GET https://api.misakey.com/identities/:id/webauthn-credentials/create
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "publicKey": {
        "challenge": "<string>",
        "rp": {
            "name": "<string>",
            "icon": "<string>",
            "id": "<string>"
        },
        "user": {
            "name": "<string>",
            "displayName": "<string>",
            "id": "<string>"
        },
        "pubKeyCredParams": [
            {
                "type": "<string>",
                "alg": <int>
            },
            [...]
        ],
        "authenticatorSelection": {
            "requireResidentKey": <bool>,
            "userVerification": "<string>"
        },
        "timeout": <int>,
        "excludeCredentials":[<array of credentials>]
    }
}
```

The response is described in the Webauthn documentation.

### Finish Webauthn credentials creation

This completes a Webauthn registration flow.

#### Request

```bash
POST https://api.misakey.com/identities/:id/webauthn-credentials/create
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_JSON Body:_
```json
{
    "name": "<string>",
    "credential": {
        "id": "<string>",
        "rawId": "<string>",
        "response": {
            "attestationObject": "<string>",
            "clientDataJSON": "<string>"
        },
        "type": "<string>"
    }
}
```

These attributes are described in the Webauthn documentation.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "id": "<base64 string>",
    "name": "<string>",
    "identity_id": "<uuid string>",
    "created_at": "<string>"
}
```

### List Webauthn Credentials

This route returns all the credentials owned by a given identity.

#### Request

```bash
GET https://api.misakey.com/webauthn-credentials?identity_id=
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Query Parameters:_
- `identity_id` (string) (uuid): the identity ID. Must be the same than the accesstoken identity id.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
    {
        "id": "<base64 string>",
        "name": "<string>",
        "identity_id": "<uuid string>",
        "created_at": "<string>"
    }
]
```

### Delete Webauthn Credential

This route deletes a given credential

#### Request

```bash
DELETE https://api.misakey.com/webauthn-credentials/:id
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the identity id owning the credential.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks

_Path Parameters:_
- `id` (string) (urlsafe base64): The credential id.

#### Response

_Code:_
```bash
HTTP 204 No Content
```

