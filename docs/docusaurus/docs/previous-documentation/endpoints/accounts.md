---
title: Accounts
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
PUT https://api.misakey.com/account/:id/password
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the account id.

_JSON Body:_
```json
{
	"old_prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"new_prehashed_password": {{% include "include/passwordHash.json" 4 %}},
	"encrypted_account_root_key": "(unpadded URL-safe base64)",
}
```

- `old_prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base_64` (base64 string).
  - `hash_base_64` (base64 string): the prehashed password.
- `new_prehashed_password` (object): prehashed password using argon2:
  - `params` (object): argon2 parameters:
    - `memory` (integer).
    - `parallelism` (integer).
    - `iterations` (integer).
    - `salt_base_64` (base64 string).
  - `hash_base_64` (base64 string): the prehashed password.
- `encrypted_account_root_key` (URL-safe base64): the account root key encrypted with the new password

### success response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Reset password

To reset password, `password_reset` must be set to true when calling `/auth/identities`.

## Get the account password parameters

This route allows the retrieval of the account password hash parameters.

Hash parameters contains information about the way the password has been hashed
following [Argon2 server relief concepts](../../concepts/server-relief/).

### Request

```bash
GET https://api.misakey.com/accounts/:id/pwd-params
```

### success response

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
- `salt_base_64` (base64 string).

## Get the account backup

This route allows the retrieval of the account backup using the unique account id.

*Note that “account secret backup” mechanism is now read-only
since the deployment of the new “secret storage” mechanism.*

### Request

```bash
GET https://api.misakey.com/accounts/:id/backup
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as an identity id linked to the account.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the unique account id.

### success response

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
