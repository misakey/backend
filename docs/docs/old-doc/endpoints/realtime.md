---
title: Box Realtime
---

# Realtime endpoints

## Getting notifications

This websocket (`wss://`) endpoint open a socket.
Notifications will be shipped through this websocket.

[More info](/old-doc/concepts/realtime.md) on the events format.

#### Request

```bash
    GET wss://api.misakey.com/box-users/74ee16b5-89be-44f7-bcdd-117f496a90a7/ws
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): the identity of the token must be the same than the one in the request
- `tokentype`: must be `bearer`

#### Response

The Websocket Protocol handshake is interpreted by HTTP servers as an Upgrade request.
The responses are similar to HTTP classic responses.

_Code_:
```bash
HTTP 200 OK
```

The websocket is then open.

# Other endpoints

## Contacting another user

This endpoint allows to create a box and invite a user
knowing their identity id and their public profile.

#### Request

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

#### Response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{
    "id": "91ec8274-2b6d-40ff-afad-83e8ba5808e5",
    "server_created_at": "2020-06-12T13:38:32.142857839Z",
    "public_key": "ShouldBeUnpaddedUrlSafeBase64",
    "title": "Test Box",
    "access_mode": "public",
    "owner_org_id": "d1e9bfa6-e931-46b1-b73c-77cb3530aadb",
    "creator": {
        "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
        "display_name": "Jean-Michel User",
        "avatar_url": null,
        "identifier_value": "jean-michel@misakey.com",
        "identifier_kind": "email"
    },
    "last_event": {
        "id": "ff6114f3-9838-40ed-a80d-bb376fd929f5",
        "type": "create",
        "content": {
            "public_key": "ShouldBeUnpaddedUrlSafeBase64",
            "title": "Test Box",
            "state": "open"
        },
        "server_event_created_at": "2020-06-12T13:38:32.142857839Z",
        "sender": {
            "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
            "display_name": "Jean-Michel User",
            "avatar_url": null,
            "identifier_value": "jean-michel@misakey.com",
            "identifier_kind": "email"
        }
    }
}
```

