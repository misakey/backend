---
title: Box Events
---

# Introduction

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.

[The shape and rules for box events are described here.](/old-doc/concepts/box-events.md)

# Events on boxes

## SINGLE creation of an event for a box

This endpoint allows the creation of events of a specific box.
Considering the type of event, side effects will occurs.

:information_source: [The shape and rules for box events are described here.](/old-doc/concepts/box-events.md)

This endpoint does not allow the creation of all type of event though. Some require to use different routes to be created as a side effect:
- `create` type events are created by the server [during the creation of the box](#creating-a-box).
- `msg.file` type events are created by the server [during the upload of an encrypted file](#upload-an-encrypted-file-to-a-box).
- `access.*` type events can only be created using [events batch creation](#batch-creation-of-events-for-a-box).
- `member.kick` type events are created by the server [during the removal of an access](/old-doc/concepts/box-events.md/#kick).

#### Request

```bash
  POST https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_JSON Body:_
```json
{
  "type": "msg.txt",
  "content": {
    "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6w+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
  },
  "referrer_id": null
}
```

#### Response

```bash
HTTP 201 Created
```

```json
{
  "id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d",
  "type": "msg.txt",
  "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
  "server_event_created_at": "2038-11-05T00:00:00.000Z",
  "sender": {
    "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
    "display_name": "Jean-Michel User",
    "avatar_url": null,
    "identifier_value": "jean-michel@misakey.com",
    "identifier_kind": "email"
  },
  "content": {
    "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6w+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
  },
  "referrer_id": null
}
```

## BATCH creation of events for a box

This endpoint allows the creation of many events in a single request, on a specific box.

This action is called a batch events creation.

:information_source: [The shape and rules for box events are described here.](/old-doc/concepts/box-events.md)

Batch event creation can't be performed using any type of events. There are type of batches.

Here is the exhaustive list of possible batch types:
- `accesses`: allow the add and removal of many accesses.

#### Request

```bash
  POST https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/batch-events
```

_JSON Body:_
```json
{
  "batch_type": "accesses",
  "events": [
    {
        "type": "access.rm",
        "referrer_id": "2c2cefaf-732c-400b-b90a-3a425a1a6d99"
    },
    {
        "type": "access.add",
        "content": {
            "restriction_type": "identifier",
            "value": "any@email.com"
        }
    }
  ]
}
```

#### Response

```bash
HTTP 201 Created
```

_JSON Body:_
```json
[
    {
        "id": "b0e8dd9f-9c0c-42b3-b00f-d92088630fd2",
        "type": "access.rm",
        "content": null,
        "referrer_id": "2c2cefaf-732c-400b-b90a-3a425a1a6d99",
        "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
        "server_event_created_at": "2020-09-14T08:06:04.054352682Z",
        "sender": {
            "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
            "display_name": "Jean-Michel User",
            "avatar_url": null,
            "identifier_value": "jean-michel@misakey.com",
            "identifier_kind": "email"
        }
    },
    {
        "id": "7807fb85-27a3-49c4-8049-e654b60c9e1d",
        "type": "access.add",
        "content": {
            "restriction_type": "identifier",
            "value": "any@email.com"
        },
        "referrer_id": null,
        "server_event_created_at": "2020-09-14T08:06:04.056930075Z",
        "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
        "sender": {
            "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
            "display_name": "Jean-Michel User",
            "avatar_url": null,
            "identifier_value": "jean-michel@misakey.com",
            "identifier_kind": "email"
        }
    },
    {
        "id": "33c05b94-3f15-4ed2-9ec9-b9c7d4c50a55",
        "type": "member.kick",
        "content": null,
        "referrer_id": "2c2cefaf-732c-400b-b90a-3a425a1a6d99",
        "server_event_created_at": "2020-09-14T08:06:04.065076788Z",
        "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
        "sender": {
          "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
          "display_name": "Jean-Michel User",
          "avatar_url": null,
          "identifier_value": "jean-michel@misakey.com",
          "identifier_kind": "email"
        }
    }
]
```

## Getting Events in a Box

#### Request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_
Pagination ([more info](/references/overview.mdx#pagination)). No pagination by default.


#### Response

_Code_:
```bash
HTTP 200 OK
```

```json
[
  (a list of events)
]
```


## Getting File Events in a Box

#### Request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_

Pagination ([more info](/references/overview.mdx#pagination)). No pagination by default.


#### Response

_Code_:
```bash
HTTP 200 OK
```

```json
[
  (a list of events of type `msg.file`)
]
```


[Events](/old-doc/concepts/box-events.md) are returned in chronological order.



## Count events for a given box

#### Request

```bash
  HEAD https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of events that the user can see.


## Count file events for a given box

#### Request

```bash
  HEAD https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of file events.
