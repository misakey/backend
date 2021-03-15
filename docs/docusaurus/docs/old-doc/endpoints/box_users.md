---
title: Box Users
---

# Contact

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
            "misakey_share": "<string> lBHT1vfwFAIBig5Nj_sD_w",
            "other_share_hash": "<string> Nz4nJMj5DOd4UGXXOlH8Ww",
            "encrypted_invitation_key_share": "<string> cGYMzgIO9rc03WoSLAyoiQdLu7he5VbMRImLhRPmwTQ"
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
# Box Settings

The box settings are unique per user settings for a given box.

We don’t manage it with events because it is specific to each user and does not describe
the box state.

It is stored in a different place and is not returned with the actual box.

## Updating a box setting

#### Request

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

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Getting a Box Setting

#### Request

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

#### Response

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
