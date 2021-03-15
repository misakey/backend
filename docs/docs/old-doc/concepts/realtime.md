---
title: Realtime
---

## Realtime

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

Each user can subscribe to `wss://api.misakey.com/box-users/:id/ws` to have realtime messages.

## Message Formats

All websockets messages are under the following format:

```json
{
    "type": "<msg type>",
    "object": {
        <content>
    }
}
```
## `event.new` server-to-client

The most important use of realtime at Misakey is to manage new box events. Their type is `event.new`.
```json
{
    "type": "event.new",
    "object": {
        <content>
    }
}
```

They are **server to client** messages.

Here are the events that can be received:

### `state.access_mode`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "state.access_mode",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "content": {
            "value": "(string) (one of: public, limited): the new access mode value of the box",
        },
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
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

### `msg.text`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "msg.text",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "content": {
            encrypted content
        },
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
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

### `msg.file`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "msg.file",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "content": {
            "encrypted_file_id": "(string) uuid of the file",
            encryption information
        },
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
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

### `msg.delete`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "msg.delete",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
        "referrer_id": "(string) uuid of the deleted message",
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
### `msg.edit`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "msg.edit",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "content": {
            new encrypted content
        },
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
        "referrer_id": "(string) uuid of the message to edit",
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

### `member.join`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "member.join",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
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

### `member.leave`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "member.leave",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
        "referrer_id": "(string) uuid of the corresponding join event",
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

### `member.kick`

```json
{
    "type": "event.new",
    "object": {
        "id": "(string) id of the event",
        "type": "member.kick",
        "box_id": "(string) id of the box",
        "owner_org_id": "(string) owner org of the box",
        "server_event_created_at": "(RFC3339 time): when the event was received by the server",
        "referrer_id": "(string) uuid of the corresponding join event",
        "sender": {
            "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
            "display_name": "Jean-Michel User",
            "avatar_url": null,
            "identifier_value": "jean-michel@misakey.com",
            "identifier_kind": "email"
        },
        "content": {
            "kicker": {
                "id": "643ca3b4-97c2-4887-b734-cb62b5bffa50",
                "display_name": "Jean-Michel Admin",
                "avatar_url": null,
                "identifier_value": "jean-michel@admin.com",
                "identifier_kind": "email"
            }
        }
    }
}
```

## Other server-to-client
### `box.delete`

This message notify a box deletion.

```json
{
    "type": "box.delete",
    "object": {
        "id": "<uuid>",
        "owner_org_id": "<uuid>",
        "sender_id": "<uuid>",
        "public_key": "<string>"
    }
}
```


### `box.settings`


This message notify a box settings update.

```json
{
    "type": "box.settings",
    "object": {
        "identity_id": "<uuid>",
        "box_id": "<uuid>",
        "owner_org_id": "<uuid>",
        "muted": "<boolean>",
    }
}
```

### `file.saved`


This message notify a change in the *saved* status of a file for a given user.

```json
{
    "type": "file.saved",
    "object": {
        "encrypted_file_id": "<uuid>",
        "is_saved": "<bool>"
    }
}
```

## Client-to-server

Server accepts only events of the type `ack`:

### `ack`

These messages are sent when a user want to acknowledge the events count on a box.

This set the events count to 0 for the user on the box.

```json
{
    "type": "ack",
    "object": {
        "sender_id": "<uuid>",
        "box_id": "<uuid>"
    }
}
```
