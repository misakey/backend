---
title: Saved Files
---

## Introduction

The saved files mecanism implements a storage space for users to keep some files on the platform.
These files must come from an existing box and can be removed from the storage space at any time.

## Creating a Saved File

This endpoints allows a user to add a file to their storage space.

#### Request

```bash
  POST https://api.misakey.com/saved-files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_JSON Body:_
```json
{
    "id": "99305633-83b4-4fd6-95b6-26987d59bc81",
    "encrypted_file_id": "596ed371-82ac-470b-938f-1fc725ccf717",
    "identity_id": "5a6f31ab-4b2c-4161-a395-c5e96933b171",
    "encrypted_metadata": "SXQgaXMgc29tZXRpbWVzIGhhcmRlciB0byByZWNlaXZlIHRoYW4gdG8gZ2l2ZS4K",
    "key_fingerprint": "e478d8300e1e1cf6e1abde3d23948e43",
    "created_at": "1994-11-05T08:15:30-05:00"
}
```

- `identity_id` (string) (uuid): id of the identity owning the saved file.
- `encrypted_file_id` (string) (uuid): id of the encrypted file.
- `encrypted_metadata` (string): encrypted metadata about the file.
- `key_fingerprint` (string): fingerprint of the key used to encrypt to file.

#### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{
    "id": "99305633-83b4-4fd6-95b6-26987d59bc81",
    "encrypted_file_id": "596ed371-82ac-470b-938f-1fc725ccf717",
    "identity_id": "5a6f31ab-4b2c-4161-a395-c5e96933b171",
    "encrypted_metadata": "SXQgaXMgc29tZXRpbWVzIGhhcmRlciB0byByZWNlaXZlIHRoYW4gdG8gZ2l2ZS4K",
    "key_fingerprint": "e478d8300e1e1cf6e1abde3d23948e43",
    "created_at": "1994-11-05T08:15:30-05:00"
}
```

## Listing the saved files

#### Request

```bash
  GET https://api.misakey.com/saved-files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Query Parameters:_
- `identity_id` (string) (uuid): a filter to list only files belonging to this identity
- Pagination ([more info](/references/overview.mdx#pagination)). No pagination by default.


#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
  {
      "id": "99305633-83b4-4fd6-95b6-26987d59bc81",
      "encrypted_file_id": "596ed371-82ac-470b-938f-1fc725ccf717",
      "identity_id": "5a6f31ab-4b2c-4161-a395-c5e96933b171",
      "encrypted_metadata": "SXQgaXMgc29tZXRpbWVzIGhhcmRlciB0byByZWNlaXZlIHRoYW4gdG8gZ2l2ZS4K",
      "key_fingerprint": "e478d8300e1e1cf6e1abde3d23948e43",
      "created_at": "1994-11-05T08:15:30-05:00"
  }
]
```

- `identity_id` (string) (uuid): id of the identity owning the saved file.
- `encrypted_file_id` (string) (uuid): id of the encrypted file.
- `encrypted_metadata` (string): encrypted metadata about the file.
- `key_fingerprint` (string): fingerprint of the key used to encrypt to file.
- `created_at` (string) (iso-8601 date): date of creation server-side

## Deleting a saved file

/!\ This won’t delete the actual file from the storage if it is linked to another entity (a box event for example).

The stored file will be deleted only in the saved file entity is the last one linked to the stored file.

#### Request

```bash
  DELETE https://api.misakey.com/saved-files/:id
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Query Parameters:_
- `id` (string) (uuid): id of the saved file to delete.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```


## Count saved files for a user

#### Request

```bash
  HEAD https://api.misakey.com/saved-files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_
- `identity_id` (string) (uuid): a filter to count only files belonging to this identity

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of file saved for a given identity_id.

## Upload directly a file to saved files

#### Request

```bash
  POST https://api.misakey.com/box-users/:id/saved-files
```

_Path Parameters:_
- `id` (uuid string): the identity id.

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Multipart Form Data Body:_
- `encrypted_metadata` (string): encrypted metadata about the file.
- `key_fingerprint` (string): fingerprint of the key used to encrypt to file.

#### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{
    "id": "99305633-83b4-4fd6-95b6-26987d59bc81",
    "encrypted_file_id": "596ed371-82ac-470b-938f-1fc725ccf717",
    "identity_id": "5a6f31ab-4b2c-4161-a395-c5e96933b171",
    "encrypted_metadata": "SXQgaXMgc29tZXRpbWVzIGhhcmRlciB0byByZWNlaXZlIHRoYW4gdG8gZ2l2ZS4K",
    "key_fingerprint": "e478d8300e1e1cf6e1abde3d23948e43",
    "created_at": "1994-11-05T08:15:30-05:00"
}
```
