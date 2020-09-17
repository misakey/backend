+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Saved Files endpoints"
tags = ["box", "saved", "files", "api", "endpoints"]
title = "Box - Saved Files"
+++

## 1. Introduction

The saved files mecanism implements a storage space for users to keep some files on the platform.
These files must come from an existing box and can be removed from the storage space at any time.

## 2. Creating a Saved File

This endpoints allows a user to add a file to their storage space.

### 2.1. request

```bash
  POST https://api.misakey.com/saved-files
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request.

_JSON Body:_
```json
{{% include "include/saved-file.json" %}}
```

- `identity_id` (string) (uuid): id of the identity owning the saved file.
- `encrypted_file_id` (string) (uuid): id of the encrypted file.
- `encrypted_metadata` (string): encrypted metadata about the file.
- `key_fingerprint` (string): fingerprint of the key used to encrypt to file.
- `created_at` (string) (iso-8601 date): date of creation server-side

### 2.2. response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{{% include "include/saved-file.json" %}}
```

## 3. Listing the saved files

### 3.1. request

```bash
  GET https://api.misakey.com/saved-files
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): the identity of the token must be the same than the one in the request.

_Query Parameters:_
- `identity_id` (string) (uuid): a filter to list only files belonging to this identity

### 3.2. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
  {{% include "include/saved-file.json" %}}
]
```

- `identity_id` (string) (uuid): id of the identity owning the saved file.
- `encrypted_file_id` (string) (uuid): id of the encrypted file.
- `encrypted_metadata` (string): encrypted metadata about the file.
- `key_fingerprint` (string): fingerprint of the key used to encrypt to file.
- `created_at` (string) (iso-8601 date): date of creation server-side

## 4. Deleting a saved file

/!\ This won’t delete the actual file from the storage if it is linked to another entity (a box event for example).

The stored file will be deleted only in the saved file entity is the last one linked to the stored file.

### 4.1. request

```bash
  DELETE https://api.misakey.com/saved-files/:id
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): the identity of the token must own the saved file to delete.

_Query Parameters:_
- `id` (string) (uuid): id of the saved file to delete.

### 4.2. response

_Code:_
```bash
HTTP 204 No Content
```
