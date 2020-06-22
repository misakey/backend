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

This route allows the creation of an account on an existing identity.

The identity must have no registered linked account.

The `prehashed_password` contains information following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
  POST https://api.misakey.com.local/identities/:id/account
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `subject` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the identity linked to the created account.

_JSON Body:_
```json
{
	"prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"backup_data": "[STRINGIFIED JSON]"
}
```

- `prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base64` (base64 string).
  - `hash_base64` (base64 string): the prehashed password.
- `backup_data` (string): the user backup data.

### Success Response

_Code:_
```bash
  HTTP 201 CREATED
```

_JSON Body:_
```json
  {
    "id": "5f80b4ec-b42a-4554-a738-4fb532ba2ee4",
    "prehashed_password": {
      "params": {{% include "include/hashParameters.json" 8 %}}
    },
    "backup_data": "[STRINGIFIED JSON]",
    "backup_version": 1,
  }
```

- `id` (uuid string): an unique id.
- `prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base64` (base64 string).
- `backup_data` (string): the stored backup data.
- `backup_version` (integer): the backup version - always 1 on creation.

## Change password

This route allows the update of an account password and the associated backup data.

The `old_prehashed_password` and `new_prehashed_password` contain information following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
  PUT https://api.misakey.com.local/account/:id/password
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 2): `subject` claim as the identity id.

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

This route allows to reset a password on an account.

The request needs to be authenticated with an ACR1 token corresponding to an identity linked to the account.

The `prehashed_password` contains information following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
  PUT https://api.misakey.com.local/account/:id/password/reset
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `subject` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the account id.

_JSON Body:_
```json
{
	"prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"backup_data": "[STRINGIFIED JSON]"
}
```

- `prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base64` (base64 string).
  - `hash_base64` (base64 string): the prehashed password.
- `backup_data` (string): the new user backup data.

### Success Response

_Code:_
```bash
  HTTP 204 NO CONTENT
```

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
- `Authorization` (opaque token) (ACR >= 2): `subject` claim as an identity id linked to the account.

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
- `Authorization` (opaque token) (ACR >= 2): `subject` claim as an identity id linked to the account.

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

### Notable Error Response

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
