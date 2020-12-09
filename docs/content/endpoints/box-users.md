+++
categories = ["Endpoints"]
date = "2020-09-28"
description = "Users endpoints"
tags = ["box", "users", "api", "endpoints"]
title = "Box - Users"
+++

# 1. Realtime endpoints

## 1.1 Getting notifications

This websocket (`wss://`) endpoint open a socket.
Notifications will be shipped through this websocket.

[More info](/concepts/realtime) on the events format.

### 1.1.1 request

```bash
    GET wss://api.misakey.com/box-users/74ee16b5-89be-44f7-bcdd-117f496a90a7/ws
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

### 1.1.2 response

The Websocket Protocol handshake is interpreted by HTTP servers as an Upgrade request.
The responses are similar to HTTP classic responses.

_Code_:
```bash
HTTP 200 OK
```

The websocket is then open.

# 2. Other endpoints

## 2.1 Contacting another user

This endpoint allows to create a box and invite a user
knowing their identity id and their public profile.

### 2.1.1 request

```bash
    POST https://api.misakey.com/box-users/74ee16b5-89be-44f7-bcdd-117f496a90a7/contact
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

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
        "extra": {
            "<string> non identified public key": "<string> encrypted crypto action"
        }
    }
```

Note that `public_key` and `other_share_hash` must be in **unpadded url-safe base64**.

When the box is created, it already contains a first event
of type `create` that contains all the information about the creation of the box.

### 2.1.2. response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```

