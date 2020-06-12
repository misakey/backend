---
title: Boxes
---

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.
[The shape and rules for box events are described here.](/concepts/box-events)


## Creating a Box

### Request

```bash
    POST https://api.misakey.com/boxes
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

_JSON Body:_
```json
    {
      "title": "Requête RGPD FNAC",
      "public_key": "SXvalkvhuhcj2UiaS4d0Q3OeuHOhMVeQT7ZGfCH2YCw"
    }
```

Where `public_key` must be in **unpadded url-safe base64**.

Note that when a box is created, it already contains a first event
of type `create` that contains all the information about the creation of the box.

### Response

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
      "lifecycle": "open",
      "creator": {{% include "include/event-sender.json" 6 %}},
      "last_event": {
        "id": "ff6114f3-9838-40ed-a80d-bb376fd929f5",
        "type": "create",
        "content": {
           "public_key": "ShouldBeUnpaddedUrlSafeBase64",
           "title": "Test Box",
           "state": "open"
        },
        "server_event_created_at": "2020-06-12T13:38:32.142857839Z",
        "sender": {{% include "include/event-sender.json" 8 %}}
      },
    }
```

The most important part is the `id` field
which must be used to interact with the box.

## Getting a Box

### Request

```bash
    GET https://api.misakey.com/boxes/:id
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

### Response

_Code:_
```bash
    HTTP 200 OK
```

_JSON Body:_
```json
    {
      "id": "91ec8274-2b6d-40ff-afad-83e8ba5808e5",
      "server_created_at": "2020-06-12T13:38:32.142857839Z",
      "public_key": "ShouldBeUnpaddedUrlSafeBase64",
      "title": "Test Box",
      "lifecycle": "open",
      "creator": {{% include "include/event-sender.json" 6 %}},
      "last_event": {
        "id": "ff6114f3-9838-40ed-a80d-bb376fd929f5",
        "type": "create",
        "content": {
           "public_key": "ShouldBeUnpaddedUrlSafeBase64",
           "title": "Test Box",
           "state": "open"
        },
        "server_event_created_at": "2020-06-12T13:38:32.142857839Z",
        "sender": {{% include "include/event-sender.json" 8 %}}
      },
    }
```

## Sending an Event to a Box

### Request

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

Note that events with type `create` cannot be posted by clients,
they are created by the backend during the creation of the box.

### Response

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


## Getting Events in a Box

### Request

```bash
    GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

### Response

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
