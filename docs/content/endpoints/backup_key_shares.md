---
title: SSO - Backup Key Shares (Key Splitting)
---
## Introduction

Key Splitting consists in splitting a secret key in several (currently, always two) *key shares*.
One share alone is completely useless, but by combining two shares of a key one can recover the secret key.

A key share has another attribute than its value,
it has an `other_share_hash` which is used for the guest frontend to identify which share it wants to retrieve.
Technically speaking, the hash is the SHA-512 hash of the other share.

## Creating a Backup Key Share

### Request

```bash
  POST https://api.misakey.com/backup-key-shares
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): the identity must be linked to an account and this account must fit the one given in the body

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

- `account_id` (string) (uuid): the account for which the shares has been created.
- `share` (string) (base64): one of the shares.
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.
- `salt_base64` (string) (base64): the salt corresponding to the backup encryption.

### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

## Getting a Backup Key Share

### Request

```bash
  GET https://api.misakey.com/backup-key-shares/:other-share-hash
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): the identity must be linked to an account and this account must fit the one for which the key has been created.

_Path Parameters:_
- `other-share-hash` (string): the hash of the key share.


### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/backup-key-share.json" %}}
```

- `account_id` (string) (uuid): the account for which the shares has been created.
- `share` (string) (base64): one of the shares.
- `other-share-hash` (string) (unpadded url-safe base64): a hash of the other share.
- `salt_base64` (string) (base64): the salt corresponding to the backup encryption.
