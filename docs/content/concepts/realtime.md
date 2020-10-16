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

Each user can subscribe to `wss://api.misakey.com/box-users/:id/ws` to have realtime messages.

## Messages

All websockets messages are under the following format:

```json
{
    "type": "<msg type>",
    "object": {
        <content>
    }
}
```

## Server to client

### `event.new` type

The most important use of realtime at Misakey is to manage new events.

Here are the events that can be received:

#### `msg.text`

```json
{
    "id": "(string) id of the event",
    "type": "msg.text",
    "box_id": "(string) id of the box",
    "content": {
        encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### `msg.file`

```json
{
    "id": "(string) id of the event",
    "type": "msg.file",
    "box_id": "(string) id of the box",
    "content": {
        "encrypted_file_id": "(string) uuid of the file",
        encryption information
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### `msg.delete`

```json
{
    "id": "(string) id of the event",
    "type": "msg.delete",
    "box_id": "(string) id of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the deleted message",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```
#### `msg.edit`

```json
{
    "id": "(string) id of the event",
    "type": "msg.edit",
    "box_id": "(string) id of the box",
    "content": {
        new encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the message to edit",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### `member.join`

```json
{
    "id": "(string) id of the event",
    "type": "member.join",
    "box_id": "(string) id of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### `member.leave`

```json
{
    "id": "(string) id of the event",
    "type": "member.leave",
    "box_id": "(string) id of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the corresponding join event",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### `member.kick`

```json
{
    "id": "(string) id of the event",
    "type": "member.kick",
    "box_id": "(string) id of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the corresponding join event",
    "sender": {{% include "include/event-identity.json" 4 %}},
    "content": {
        "kicker": {{% include "include/event-identity.json" 8 %}}
    }
}
```

## Notifications

### Server to Client

Notifications object are:

```json
{
    "type": "<type>",
    "object": {
        <object>
    }
}
```

### `box.delete` type


This message notify a box deletion.

```json
{
    "id": "<uuid>",
    "sender_id": "<uuid>",
    "public_key": "<string>"
}
```

## Client to server

Server accepts only events of the type `ack`:

#### `ack` type

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
