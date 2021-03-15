---
title: Boxes
---

# Introduction

A box is a space where the end-identity can share securely the data with some other identity or themself.

It is the base for data exchange, data access management...

# Boxes

## Creating a Box

### Request

```bash
  POST https://api.misakey.com/boxes
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): no identity check, just a valid token is required.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_JSON Body:_
```json
    {
      "title": "Requête RGPD",
      "owner_org_id": "d1e9bfa6-e931-46b1-b73c-77cb3530aadb",
      "datatag_id": "b7073bc5-b2e8-4a22-9717-8418de13bfa5",
      "data_subject": "michel@misakey.com",
      "public_key": "SXvalkvhuhcj2UiaS4d0Q3OeuHOhMVeQT7ZGfCH2YCw",
      "key_share": {
        "misakey_share": "lBHT1vfwFAIBig5Nj_sD_w",
        "other_share_hash": "Nz4nJMj5DOd4UGXXOlH8Ww",
        "encrypted_invitation_key_share": "cGYMzgIO9rc03WoSLAyoiQdLu7he5VbMRImLhRPmwTQ"
      }
      "invitation_data": {
        "<public key>": "encrypted crypto action"
       }
    }
```

- `title` is a free text required that is meant to describe the box purpose.
- `owner_org_id` is an optional uuid corresponding to the organization owning the box (default is self org).
- `datatag_id` is an **optional** uuid corresponding to a datatag representing the data type shipped through this box. `owner_org_id` must also bit set.
- `datatag_subject` is an **optional** identifier to define the data subject of this box.
- `public_key` and `other_share_hash` must be in **unpadded url-safe base64**. If the public key is for `com.misakey.aes-rsa-enc` algorithm, it must be prefixed with `com.misakey.aes-rsa-enc:`.
- `key_share` is **optional** but will be soon mandatory.
- `invitation_data` is **optional** but must be used if the `datatag_subject` has an account. It contains the encrypted crypto action to directly invite the user (see the `extra` field in the [access events section](../../concepts/box-events/#2511-to-a-specific-identifier-value))

When a box is created, it already contains a first event
of type `create` that contains most of the information about the creation of the box.

Note that the access mode of a box is limited by default. A `state.access_mode` event must be created to switch it to `public`, see the related event type documentation.

### Response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```

The most important part is the `id` field
which must be used to interact with the box.

## Getting a Box

### Request

```bash
  GET https://api.misakey.com/boxes/:id
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): no identity check, just a valid token is required.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/box-read.json" %}}
```

### notable error reponses

**I - The identity has no access to the box

The reason of the forbidden is explained is a reason field that have only 2 possible fixed values.
Only the get box endpoint is ensured to return this error in the current state of the API.

```bash
HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
    "code": "forbidden",
    "origin": "not_defined",
    "desc": "",
    "details": {
        "reason": "no_access|not_member"
    }
}
```

## Getting public info about a box

This endpoint does not need any valid token, but it needs a valid `other_share_hash`
corresponding to the box to get.

### Request

```bash
  GET https://api.misakey.com/boxes/:id/public?other_share_hash=
```

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

_Query Parameters:_
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.

### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "title": "<title of the box>",
    "owner_org_id": "<uuid representing the organization owning the box>",
    "creator": {{% include "include/event-identity.json" %}}
}
```
## Delete a box

[Box admins](../../concepts/box-events/#21-admins) only are able to delete corresponding boxes.

A removed box sees its data completely removed from Misakey storage. This action is irreversible.

This action removes all data related to the box (events, key-shares...).

### Request

```bash
DELETE https://api.misakey.com/boxes/:id
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): the linked identity must be considered as a [box admin](../concepts/box-events/#21-admins).
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the box id wished to be deleted.

_JSON Body:_
```json
{
  "user_confirmation": "delete"
}
```

- `user_confirmation` (string) (one of: _delete_, _supprimer_): the input the end-identity has entered to confirm the deletion. The server will check if the value corresponds to some expected strings (cf one of).

### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Reset the new events count for an identity

The list of boxes return many information for each boxes, including a numerical field `events_count` telling how many event have occured since the connected identity's last visit.

This endpoint allows to reset the new events count of a box for a given identity.

It is a kind of an acknowledgement and it must be used when the identity want to mark the box as "read".

### Request

```bash
PUT https://api.misakey.com/boxes/:id/new-events-count/ack
```

_Path Parameter:_
- `id` (string) (uuid): the box id to mark as "read"

_JSON Body:_
```json
    {
        "identity_id": "e2e49259-f840-4991-a9f7-97c5f267bd18"
    }
```

where `identity_id` is the identity of the requester who wants to acknowledge.

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid access token corresponding to the identity of the body
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

### success response

_Code:_
```bash
HTTP 204 NO CONTENT
```

# Accesses

Access defines who has access to a box considering some rules.

## Add or remove accesses

:warning: Only box admins can add or remove accesses.

To add or remove accesses in a given box, please refer to:
* the [sending an event to a box endpoint](../box_events/#21-single-creation-of-an-event-for-a-box).
* the [access type events documentation](/concepts/box-events/#15-access-type-events).

## Listing accesses for a given box

Listing accesses allows admins to see the current state of the box reachability.

### Request

```bash
GET https://api.misakey.com/boxes/:id/accesses
```

_Path Parameter:_
- `id` (string) (uuid): the box id to list accesses on.

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the access token should belongs the a box admin.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

### success response

Only the current valid accesses are returned.

```json
[
    {
      "id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d",
      "type": "access.add",
      "server_event_created_at": "2038-11-05T00:00:00.000Z",
      "content": {
          "restriction_type": "email_domain",
          "value": "misakey.com"
      }
    },
    {
      "id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d",
      "type": "access.add",
      "server_event_created_at": "2038-11-05T00:00:00.000Z",
      "content": {
          "restriction_type": "email",
          "value": "sadin.nicolas7@gmail.com"
      }
    }
]
```

# Membership

## Add or remove membership

To add or remove membership for a couple <box, identity>, please refer to:

* the [sending an event to a box endpoint](../box_events/#21-single-creation-of-an-event-for-a-box).
* the [member type events documentation](/concepts/box-events/#12-member-type-events).

## List boxes for the current identity

### Request

Users are able to list boxes they have an access to.

The returned list is automatically computed from the server according to the authorization
provided by the received bearer token.

```bash
GET https://api.misakey.com/boxes/joined
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_
- `owner_org_id` (uuid) (default: hosting-org): the organization id owning the box.
- `datatag_id` (uuid): the datatag id corresponding to the data type. `""` for boxes with no datatag. Empty parameter for all boxes.
- Pagination ([more info](/concepts/pagination)) with default limit set to 10.

### Response

_Code:_
```bash
HTTP 200 OK
```

A list of event is returned.
```json
[
  {{% include "include/box-read.json" 2 %}}
]
```

## Count boxes for the current identity

### Request

This request allows to retrieval of information about accessible boxes list.

Today only the total count of boxes is returned as an response header.

```bash
  HEAD https://api.misakey.com/boxes/joined?owner_org_id=d1e9bfa6-e931-46b1-b73c-77cb3530aadb
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_
- `owner_org_id` (uuid) (default: hosting-org): the organization id owning the box.
- `datatag_id` (uuid): the datatag id corresponding to the data type. `""` for boxes with no datatag. Empty parameter for all boxes.

### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of boxes that the identity can access.

## Listing active members of a box

This endpoint return all identities that have an active membership to the box.

Identities who have left and have been kicked out of the box are not returned.

### Request

```bash
GET https://api.misakey.com/boxes/:id/members
```

_Path Parameter:_
- `id` (string) (uuid): the box id

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

### Response

_Code:_
```bash
HTTP 200 OK
```

A list of senders is returned.
```json
[
  {{% include "include/event-identity.json" %}}
]
```
