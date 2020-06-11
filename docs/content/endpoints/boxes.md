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

    POST https://api.misakey.com/boxes

    {
      "public_key": "SXvalkvhuhcj2UiaS4d0Q3OeuHOhMVeQT7ZGfCH2YCw",
      "title": "Requête RGPD FNAC",
    }

Where `public_key` is the public key of the box,
the key that must be used to encrypt messages sent to this box,
and `title` is the title of the box.

Note that when a box is created, it already contains a first event
of type `create` that contains all the information about the creation of the box.

### Response

    HTTP 201 Created

    {
      "public_key": "SXvalkvhuhcj2UiaS4d0Q3OeuHOhMVeQT7ZGfCH2YCw",
      "title": "Requête RGPD FNAC",
      "id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
      "creator": {{% include "include/event-sender.json" 6 %}},
      "server_created_at": "2020-04-01T20:22:45.691Z"
    }

The most important part is the `id` field
which must be used to interact with the box.


## Getting Events in a Box

    GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events

### Response

    HTTP 200 OK

    [
      (a list of events)
    ]

Events are returned in chronological order.


## Sending an Event to a Box

    POST https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events

    {
      "type": "msg.txt",
      "content": {
        "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6wLV+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
      }
    },

Note that events with type `create` cannot be posted by clients,
they are created by the backend during the creation of the box.

### Response

    HTTP 201 Created

    {
      "id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d",
      "type": "msg.txt",
      "server_event_created_at": "2020-04-01T20:22:45.691Z",
      "sender": {{% include "include/event-sender.json" 6 %}},
      "content": {
        "encrypted": "UrxdLg+Z5cyeRMz8/zk2aKxRlW9jwKf9FPskm8QO8EeiSm3B+Hj3JbvTdCnbsLVB8bjVC/GHYuzabHogpbXNuBTiFSMau3G81OkSoLDo58q6X8Rq7PE/ULcHhB1sClJ63Qk5DyTOXSPA3yr2LQTY0gfKLSnAT45H3d6wLV+fg5LEAtsJV3hRAZfiKd0dRjv7UZxS4rUAr2BM5EDA2lGP4az8Vd9xyhSmYiNPPDXEWwBmFFSUM8PaA9Lnectl2VjLLY4mDmhbjnBF+9WntV42Baa4zfP46Zxhq1EbGjPItStWPSZl4onKg1BUP2qcHQBqjoliIiuru7rw3Qd/7zse8A=="
      }
    },


