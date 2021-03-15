+++
categories = ["Concepts"]
date = "2020-09-15"
description = "Realtime Management"
tags = ["concepts", "realtime", "websockets"]
title = "Realtime"
+++

# 1. Realtime

At Misakey, realtime is managed through **websockets** or **polling**.
We aim at manage our whole realtime through **websockets**.

## 1.1. Websockets

The websockets protocol uses the same handshake as http requests.
Servers can differenciate them thanks to the `Connection: Upgrade` header.

Our server send regular (every 60 seconds) `Pings` to check the connection state.
For now, there is no action triggered by a lack of a `Pong` response.

Authentication is made through an **access token**. 
As the javascript lib does not allow custom headers, we need to
pass the access token through query parameters.
We will improve this authentication process in the future.

Each user can subscribe to `wss://api.misakey.com/box-users/:id/ws` to have realtime messages.

## 1.2. Messages

All websockets messages are under the following format:

```json
{
    "type": "<msg type>",
    "object": {
        <content>
    }
}
```

### 1.2.1. Server to client

### 1.2.2. `event.new` type

The most important use of realtime at Misakey is to manage new events.

Here are the events that can be received:

#### 1.2.2.1. `state.access_mode`

```json
{
    "id": "(string) id of the event",
    "type": "state.access_mode",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "content": {
        "value": "(string) (one of: public, limited): the new access mode value of the box",
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.2. `msg.text`

```json
{
    "id": "(string) id of the event",
    "type": "msg.text",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "content": {
        encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.3. `msg.file`

```json
{
    "id": "(string) id of the event",
    "type": "msg.file",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "content": {
        "encrypted_file_id": "(string) uuid of the file",
        encryption information
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.4. `msg.delete`

```json
{
    "id": "(string) id of the event",
    "type": "msg.delete",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the deleted message",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```
#### 1.2.2.5. `msg.edit`

```json
{
    "id": "(string) id of the event",
    "type": "msg.edit",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "content": {
        new encrypted content
    },
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the message to edit",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.6. `member.join`

```json
{
    "id": "(string) id of the event",
    "type": "member.join",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.7. `member.leave`

```json
{
    "id": "(string) id of the event",
    "type": "member.leave",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the corresponding join event",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

#### 1.2.2.8. `member.kick`

```json
{
    "id": "(string) id of the event",
    "type": "member.kick",
    "box_id": "(string) id of the box",
    "owner_org_id": "(string) owner org of the box",
    "server_event_created_at": "(RFC3339 time): when the event was received by the server",
    "referrer_id": "(string) uuid of the corresponding join event",
    "sender": {{% include "include/event-identity.json" 4 %}},
    "content": {
        "kicker": {{% include "include/event-identity.json" 8 %}}
    }
}
```

## 1.3. Notifications

### 1.3.1. Server to Client

Notifications object are:

```json
{
    "type": "<type>",
    "object": {
        <object>
    }
}
```

### 1.3.2. `box.delete` type


This message notify a box deletion.

```json
{
    "id": "<uuid>",
    "owner_org_id": "<uuid>",
    "sender_id": "<uuid>",
    "public_key": "<string>"
}
```


### 1.3.3. `box.settings` type


This message notify a box settings update.

```json
{ 
    "identity_id": "<uuid>",
    "box_id": "<uuid>",
    "owner_org_id": "<uuid>",
    "muted": "<boolean>",
}
```

### 1.3.4. `file.saved` type


This message notify a change in the *saved* status of a file for a given user.

```json
{
    "encrypted_file_id": "<uuid>",
    "is_saved": "<bool>"
}
```

## 1.4. Client to server

Server accepts only events of the type `ack`:

#### 1.4.0.1. `ack` type

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
