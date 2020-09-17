+++
categories = ["Concepts"]
date = "2020-09-15"
description = "Realtime Management"
tags = ["concepts", "realtime", "websockets"]
title = "Realtime"
+++

# Realtime

At Misakey, realtime is managed through **websockets** or **polling**.
We aim at manage our whole realtime through **websockets**.

## Websockets

The websockets protocol uses the same handshake as http requests.
Servers can differenciate them thanks to the `Connection: Upgrade` header.

Our server send regular (every 60 seconds) `Pings` to check the connection state.
For now, there is no action triggered by a lack of a `Pong` response.

Authentication is made through an **access token**. 
As the javascript lib does not allow custom headers, we need to
pass the access token through query parameters.
We will improve this authentication process in the future.

## Events

The most important use of realtime at Misakey is to manage new events.
Events are shipped through websockets thanks to the `wss://api.misakey.com/boxes/{id}/events/ws` endpoint.

Here are the events that can be received on this socket:

### `msg.text`

```json
{
    "id": "(string) id of the event",
    "type": "msg.text",
    "content": {
        encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    }
}
```

### `msg.file`

```json
{
    "id": "(string) id of the event",
    "type": "msg.file",
    "content": {
        "encrypted_file_id": "(string) uuid of the file",
        encryption information
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    }
}
```

### `msg.delete`

```json
{
    "id": "(string) id of the event",
    "type": "msg.delete",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    },
    "referrer_id": "(string) uuid of the deleted message"
}
```
### `msg.edit`

```json
{
    "id": "(string) id of the event",
    "type": "msg.edit",
    "content": {
        new encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    },
    "referrer_id": "(string) uuid of the message to edit"
}
```

### `member.join`

```json
{
    "id": "(string) id of the event",
    "type": "member.join",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    },
}
```

### `member.leave`

```json
{
    "id": "(string) id of the event",
    "type": "member.leave",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    },
    "referrer_id": "(string) uuid of the corresponding join event"
}
```

### `member.kick`

```json
{
    "id": "(string) id of the event",
    "type": "member.kick",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {
      "display_name": "(string) the display name of the sender",
      "avatar_url": "(string) (nullable) the potential avatar url of the sender",
      "identifier": {
        "value": "(string) the value of the identifier",
        "kind": "(string) (one of: email): the kind of the identifier"
      }
    },
    "referrer_id": "(string) uuid of the corresponding join event"
}
```
