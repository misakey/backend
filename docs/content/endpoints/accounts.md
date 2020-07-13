---
title: SSO - Accounts
---

## Introduction

**Account** is an entity within the system representing theorically one physical person
in the "real world".

It is used for 3 main reasons:
- link [identities](../identities/) together (one person can have many identities).
- store a password (for authentication flow and for the use cryptographic features).
- store some backup data (data used to make cryptographic features better).

An account has always an identity linked to it, it cannot exist alone. Though it is
important to notice it is identities that contains that link information, considering the one (account)
to many (identities) relationship.

## Create an account on an identity

The creation of an account linked to an identity can be done in an auth flow.

More information in the [auth flow section](../auth_flow/#method-name-account_creation-bust_in_silhouette).

## Change password

This route allows the update of an account password and the associated backup data.

The `old_prehashed_password` and `new_prehashed_password` contain information following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
PUT https://api.misakey.com.local/account/:id/password
```
_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 2): `subject` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the account id.

_JSON Body:_
```json
{
	"old_prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"new_prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"backup_data": "[STRINGIFIED JSON]",
    "backup_version": 3
}
```

- `old_prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base64` (base64 string).
  - `hash_base64` (base64 string): the prehashed password.
- `new_prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base64` (base64 string).
  - `hash_base64` (base64 string): the prehashed password.
- `backup_data` (string): the new user backup data.
- `backup_version` (integer): the new backup data version (must be current version + 1).

### Success Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Reset password

The reset password is possible as an extension of an authentication step within an auth flow.

[See here for more information](../auth_flow/#reset-password-extension)

## Get the account password parameters

This route allows the retrieval of the account password hash parameters.

Hash parameters contains information about the way the password has been hashed
following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
GET https://api.misakey.com.local/accounts/:id/pwd-params
```

_Headers:_
- No `Authorization` is required to retrieve the resource.

### Success Response

_Code:_
```bash
HTTP 200 OK
```

```json
{{% include "include/hashParameters.json" %}}
```

- `memory` (integer).
- `parallelism` (integer).
- `iterations` (integer).
- `salt_base64` (base64 string).

## Get the account backup

This route allows the retrieval of the account backup using the unique account id.

### Request

```bash
GET https://api.misakey.com.local/accounts/:id/backup
```
_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 2): `subject` claim as an identity id linked to the account.

_Path Parameters:_
- `id` (uuid string): the unique account id.

### Success Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "data": "[STRINGIFIED JSON]",
    "version": 3
}
```

- `data` (string): the user backup data.
- `version` (integer): the current backup version.

____
## Update the account backup

This route allows the update of the account backup using the unique account id.

### Request

```bash
PUT https://api.misakey.com.local/accounts/:id/backup
```
_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 2): `subject` claim as an identity id linked to the account.

_Path Parameters:_
- `id` (uuid string): the unique account id.

_JSON Body:_
```json
{
    "data": "[STRINGIFIED JSON]",
    "version": 3
}
```

- `data` (string): the user backup data.
- `version` (integer): this value is expected to be equal to 1 + the version of the backup currently stored.
The client informs the server it increase the version number by updating the backup data.

### Success Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

### Notable Error Responses

On errors, some information should be displayed to the end-user.

**1. Received version invalid:**

This error occurs when the received new version is refused by the server.
Either the received version is too low or too high.

An "expected_version" field is present in details to inform which version number
is expected from the server considering the current backup version.

_Code:_
```bash
HTTP 409 CONFLICT
```

_JSON Body:_
```json
{
    "code": "conflict",
    "origin": "body",
    "details": {
        "version": "conflict",
        "expected_version": "5"
    }
}
