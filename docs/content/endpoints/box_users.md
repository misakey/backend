+++
categories = ["Endpoints"]
date = "2020-12-31"
description = "Box Users Endpoints"
tags = ["box", "users", "api", "endpoints"]
title = "Box Users"
+++

# 1. Contact

## 1.1 Contacting another user

This endpoint allows to create a box and invite a user
knowing their identity id and their public profile.

### 1.1.1 request

```bash
    POST https://api.misakey.com/box-users/74ee16b5-89be-44f7-bcdd-117f496a90a7/contact
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_JSON Body:_
```json
    {
        "contacted_identity_id": "<string> 0f10d6a5-0626-4492-8e4c-2c37b258f921 ",
        "box": {
            "title": "<string> User1 - User2",
            "public_key": "<string> SXvalkvhuhcj2UiaS4d0Q3OeuHOhMVeQT7ZGfCH2YCw",
        }
        "key_share": {
            "misakey_share": "<string> lBHT1vfwFAIBig5Nj+sD+w==",
            "other_share_hash": "<string> Nz4nJMj5DOd4UGXXOlH8Ww",
            "encrypted_invitation_key_share": "<string> cGYMzgIO9rc03WoSLAyoiQdLu7he5VbMRImLhRPmwTQ="
        },
        "invitation_data": {
            "<string> non identified public key": "<string> encrypted crypto action"
        }
    }
```

Note that `public_key` and `other_share_hash` must be in **unpadded url-safe base64**.

When the box is created, it already contains a first event
of type `create` that contains all the information about the creation of the box.

### 1.1.2. response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```
# 2. Box Settings

The box settings are unique per user settings for a given box.

We don’t manage it with events because it is specific to each user and does not describe
the box state.

It is stored in a different place and is not returned with the actual box.

## 2.1 Updating a box setting

### 2.1.1. request

```bash
  PUT https://api.misakey.com/box-users/:id/boxes/:bid/settings
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): identity must be the same than the requested identity.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (string): the identity id
- `bid` (string): the box id

_JSON Body:_
```json
{
    "muted": true|false
}
```

- `muted` (bool): is the user notified on a box update;

### 2.1.2 response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## 2.2. Getting a Box Setting

### 2.2.1 request

```bash
  GET https://api.misakey.com/box-users/:id/boxes/:bid/settings
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): identity must be the same than the requested identity.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (string): the identity id
- `bid` (string): the box id

### 2.2.2 response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "identity_id": "41e213d1-7d85-4b08-b913-678da2653021",
    "box_id": "89e213d1-7d85-4b08-b913-678da2653846",
    "muted": true|false
}
```

- `identity_id` (string) (uuid): the identity id.
- `box_id` (string) (uuid): the box id.
- `muted` (bool):is the user notified on a box update?
