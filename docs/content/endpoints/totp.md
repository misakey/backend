+++
categories = ["Endpoints"]
date = "2021-02-04"
description = "TOTP endpoints"
tags = ["sso", "totp", "secret", "api", "endpoints"]
title = "SSO - TOTP"
+++

## 1. Introduction

### 1.1. Concept

TOTP is used in our **2FA** mechanism.

It must register on a external TOTP App via a QR Code.

## 2. TOTP

### 2.1 Delete TOTP Secret

This route deletes the unique identity secret.

The identity **must not** have `mfa_method` configured to `totp`.

#### 2.1.1 request

```bash
DELETE https://api.misakey.com/identities/:id/totp
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the identity id owning the credential.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks

_Path Parameters:_
- `id` (string) (uuid): The identity id.

#### 2.1.2. success response

_Code:_
```bash
HTTP 204 No Content
```

### 2.2. Configure TOTP

This initiates a TOTP Enrollment.

#### 2.2.1. request

```bash
GET https://api.misakey.com/identities/:id/totp/enroll
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### 2.2.2. success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "id": "<uuid string>",
    "base64_image": "<b64 image>"
}
```

- `id`: an uuid to identify the enrollment flow
- `base64_image`: the QR code image encoded in base64

### 2.3. Finish TOTP enrollment

This completes a TOTP Enrollment flow.

#### 2.3.1. request

```bash
POST https://api.misakey.com/identities/:id/totp/enroll
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
    "id": "<uuid string>",
    "code": "<string>"
}
```

- `id`: the unique id identifying the enrollment flow
- `code`: the code returned by the external app when registering via the QR code

#### 2.3.2. success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "recovery_codes": [
        "<string>",
        "<string>"
    ]
}
```

- `recovery_codes`: a set of one time use codes that can be used instead of the code during auth flow

### 2.4. Regenerate recovery codes

This allows a user to regenerate their set of recovery codes.

It erases the old set.

#### 2.4.1. request

```bash
POST https://api.misakey.com/identities/:id/totp/recovery-codes
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 3): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### 2.3.2. success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "recovery_codes": [
        "<string>",
        "<string>"
    ]
}
```

- `recovery_codes`: a set of one time use codes that can be used instead of the code during auth flow


