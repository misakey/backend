+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Events endpoints"
tags = ["box", "events", "api", "endpoints"]
title = "Box - Events"
+++

# 1. Introduction

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.

[The shape and rules for box events are described here.](/concepts/box-events)

# 2. Events on boxes

## 2.1. SINGLE creation of an event for a box

This endpoint allows the creation of events of a specific box.
Considering the type of event, side effects will occurs.

:information_source: [The shape and rules for box events are described here.](/concepts/box-events)

This endpoint does not allow the creation of all type of event though. Some require to use different routes to be created as a side effect:
- `create` type events are created by the server [during the creation of the box](../boxes/#21-creating-a-box).
- `msg.file` type events are created by the server [during the upload of an encrypted file](../box_enc_files/#23-upload-an-encrypted-file-to-a-box).
- `access.*` type events can only be created using [events batch creation](./#22-batch-creation-of-events-for-a-box).
- `member.kick` type events are created by the server [during the removal of an access](../../concepts/box-events/#123-kick).

### 2.1.1. request

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

### 2.1.2. response

```bash
HTTP 201 Created
```

```json
{
  "id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d",
  "type": "msg.txt",
  "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
  "server_event_created_at": "2020-04-01T20:22:45.691Z",
  "sender": {{% include "include/event-identity.json" 2 %}},
  "content": {
    "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6w+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
  },
  "referrer_id": null
}
```

### 2.1.3. notable error reponses

**I - The box is closed:**

A box that has received a lifecycle closed event cannot received new events.

```bash
HTTP 409 CONFLICT
```

_JSON Body:_
```json
{
    "code": "conflict",
    "origin": "not_defined",
    "desc": "box is closed.",
    "details": {
        "lifecycle": "conflict"
    }
}
```

## 2.2. BATCH creation of events for a box

This endpoint allows the creation of many events in a single request, on a specific box.

This action is called a batch events creation.

:information_source: [The shape and rules for box events are described here.](/concepts/box-events)

Batch event creation can't be performed using any type of events. There are type of batches.

Here is the exhaustive list of possible batch types:
- `accesses`: allow the add and removal of many accesses.

### 2.2.1. request

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

### 2.2.2. response

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
        "sender": {{% include "include/event-identity.json" 8 %}}
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
        "sender": {{% include "include/event-identity.json" 8 %}}
    },
    {
        "id": "33c05b94-3f15-4ed2-9ec9-b9c7d4c50a55",
        "type": "member.kick",
        "content": null,
        "referrer_id": "2c2cefaf-732c-400b-b90a-3a425a1a6d99",
        "server_event_created_at": "2020-09-14T08:06:04.065076788Z",
        "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
        "sender": {{% include "include/event-identity.json" 8 %}}
    }
]
```

### 2.2.3. notable error reponses

**I - The box is closed:**

A box that has received a lifecycle closed event cannot received new events.

```bash
HTTP 409 CONFLICT
```

_JSON Body:_
```json
{
    "code": "conflict",
    "origin": "not_defined",
    "desc": "box is closed.",
    "details": {
        "lifecycle": "conflict"
    }
}
```

## 2.3. Getting Events in a Box

### 2.3.1. request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

_Query Parameters:_

Pagination ([more info](/concepts/pagination)). No pagination by default.


### 2.3.2. response

_Code_:
```bash
HTTP 200 OK
```

```json
[
  (a list of events)
]
```


## 2.4 Getting File Events in a Box

### 2.4.1. request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

_Query Parameters:_

Pagination ([more info](/concepts/pagination)). No pagination by default.


### 2.4.2. response

_Code_:
```bash
HTTP 200 OK
```

```json
[
  (a list of events of type `msg.file`)
]
```


[Events](/concepts/box-events) are returned in chronological order.



## 2.5 Get the total count of events for a given box

### 2.5.1. request

```bash
  HEAD https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

### 2.5.2. response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of events that the user can see.


## 2.6 Get the total count of file events for a given box

### 2.6.1. request

```bash
  HEAD https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/files
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): a valid token.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

### 2.6.2. response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of file events.
