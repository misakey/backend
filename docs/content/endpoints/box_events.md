---
title: Box - Events
---

# 1. Introduction

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.

[The shape and rules for box events are described here.](/concepts/box-events)

# 2. Events on boxes

## 2.1. Sending an Event to a Box

This endpoint allows the creation of events of a specific box.
Considering the type of event, side effects will occurs.

:information_source: [The shape and rules for box events are described here.](/concepts/box-events)

This endpoint does not allow the creation of all type of event though. Some require to use different routes to be created as a side effect:
- `create` type events are created by the server [during the creation of the box](../boxes/#21-creating-a-box).
- `msg.file` type events are created by the server [during the upload of an encrypted file](../box_enc_files/#23-upload-an-encrypted-file-to-a-box).
- `member.join` type events are created by the server [during the access with a key share](../box_key_shares/#3-getting-a-box-key-share)

### 2.1.1. request

```bash
  POST https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_JSON Body:_
```json
    {
      "type": "msg.txt",
      "content": {
        "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6wLV+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
      }
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
      "server_event_created_at": "2020-04-01T20:22:45.691Z",
      "sender": {{% include "include/event-sender.json" 6 %}},
      "content": {
        "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6wLV+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
      }
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
        "lifecyle": "conflict"
    }
}
```

## 2.2. Getting Events in a Box

### 2.2.1. request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): a valid token.

_Query Parameters:_

Pagination ([more info](/concepts/pagination)). No pagination by default.


### 2.2.2. response

_Code_:
```bash
HTTP 200 OK
```

```json
    [
      (a list of events)
    ]
```

[Events](/concepts/box-events) are returned in chronological order.
