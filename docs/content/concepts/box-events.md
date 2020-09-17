+++
categories = ["Concepts"]
date = "2020-09-11"
description = "Box Events"
tags = ["concepts", "box", "events"]
title = "Box Events"
+++

# 1. Events

*Boxes* objects contain *events*.
An event represents the messages sent to the box
as well as changes made to the state of the box
(closing the box, and in the future other operations such as changing the title etc ...).

Events are mainly used in the endpoints described [here](/endpoints/boxes) and [here](/endpoints/box_events).

An event has the following fields:

```json
{
  "id": "(string) (uuid): unique id set by the server",
  "server_event_created_at": "(RFC3339 time): when the event was received by the server",
  "sender": {
    "display_name": "(string) the display name of the sender",
    "avatar_url": "(string) (nullable) the potential avatar url of the sender",
    "identifier": {
      "value": "(string) the value of the identifier",
      "kind": "(string) (one of: email): the kind of the identifier"
    }
  },
  "type": "(string) (one of: create, msg.txt, msg.file, state.lifecycle, member.join, member.leave): the type of the event",
  "content": "(json object) (nullable): its shape depends on the type of event - see definitions below",
  "referrer_id": "(string) (uuid) (nullable): the uuid of a potential referrer event"
}
```

## 1.1. `Create` type event

A box starts with a `create` event
which is not posted by the client
but is instead inserted automatically by the backend
at the creation of the box.

An event of type `create` has the following content fields:

```json
{
  "type": "create",
  "content": {
    "public_key": "(string): the public key that must be used to encrypt messages for this box",
    "title": "(string): the title of the box"
  },
  "referrer_id": null
}
```

## 1.2. `Member` type events

### 1.2.1. Join

The `member.join` event is sent by the user when they want to be a member of the box.
The user must have access to the box to send such an event.

Messages of type `member.join` have no content field:

```json
{
  "type": "member.join",
  "content": null,
  "referrer_id": null
}
```

### 1.2.2. Leave

An event of type `member.leave` can be added by the user if they are member of the box.
It will automatically refers the previous `member.join` event of the user.
An admin can not create a `member.leave` event on their box.

```json
{
  "type": "member.leave",
  "referrer_id": "<member.join id> (automatically added by the server)",
  "content": null
}
```

### 1.2.3. Kick

An event of type `member.kick` is automatically added to the box by the backend
when an `access.rm` removes the last access rules allowing some members to access the box.

If any other rules allows the members to access the box, the event is not triggered.

The event refers the previous `member.join` event of the user, the `sender_id` is set to the person who has changed access rules and has then triggered this event.

Messages of type `member.kick` have no content.

```json
{
  "type": "member.kick",
  "referrer_id": "(string) (uuid): member.join id",
  "content": {
      "kicked_member_id": "(string) (uuid): identity id that has been kicked"
  },
}
```

Once the event has been created for a user, the user is still listed as member but informed that he has been kicked.

On read, the `kicked_member_id` is transformed into a `kicked_member` field containing sender information. This `kicked_member` attribute is nullable.

## 1.3. `Message` type events

Messages allow the transfer of encrypted message text or blob data.

### 1.3.1. Message Text
Messages of type `msg.txt` allow the transfer of message text.

```json
{
  "type": "msg.txt",
  "content": {
    "encrypted": "(string) (base64): the encrypted message. Recall that files are sent separately from the message, so the size of a message event stays rather small.",
    "deleted": { // nullable, indicates the message have been removed
      "at_time": "indicates the deletion time of the message",
      "by_identity": "indicates who has deleted the message",
    },
    "last_edited_at": "(RFC3339 time): indicates that the message was edited, and when",
    "referrer_id": null
}
```

### 1.3.2. Message File

Messages of type `msg.file` allow the transfer of blob data.

```json
{
  "type": "msg.file",
  "content": {
    "encrypted": "(string) (base64): information about file encryption.",
    "encrypted_file_id": "(string) (uuid format): a unique identifier used to store and download the file",
    "deleted": { // nullable, indicates the message have been removed
      "at_time": "indicates the deletion time of the message",
      "by_identity": "indicates who has deleted the message",
    },
    "referrer_id": null
}
```

### 1.3.3. Deleting a Message Event

A message (text or file) can be deleted by its author or by the box admin.

The event is still present in the box but the encrypted content is removed
and replaced by who deleted it and when.

```json
{
  "type": "msg.delete",
  "referrer_id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d"
}
```

Where `referrer_id` is the ID of the event to delete.

- The sender's account must be the one that sent the event to delete,
  or the sender must be the box creator.
- the message must not be already deleted
- the box must be not be closed

### 1.3.4. Editing a Message

Users can edit their own messages.

```json
{
    "type": "msg.edit",
    "content": {
        "new_encrypted": "EditedXXB64dcc9PhJTeyUS2K04zeHKLMW8fviUkmyBjWdGvwwo=",
        "new_public_key": "EditedXXa75RO1FzZpskiKHAggyB7YNJoz4R24dnMFvHfMzu4wQ="
    },
    "referrer_id": "7410feae-637e-40a8-ab59-badeaf479c63"
}
```

Where `referrer_id` is the ID of the event to edit.

- The sender's account must be the one that sent the event to edit.
- the message must not be already deleted
- the box must be not be closed

## 1.4. `State` type events

State events are meant to update the state of a box: information that can be set to only one value and are overriden when updated.

### 1.4.1. `State Lifecycle`

The `state.lifecycle` event changes the lifecycle of a box.

```json
{
  "type": "state.lifecycle",
  "content": {
    "state": "(string) (one of: closed): the new state of the box"
  },
  "referrer_id": null
}
```

## 1.5. `Access` type events

Access events are specific rules defined by the admins and allowing considering their logic who can access the box.

### 1.5.1. Add

The add of an access is represented by this event shape
```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (one of : invitation_link, identifier or email_domain): the type of restriction the access bears",
        "value": "(string): a value describing the restriction"
    },
    "referrer_id": null
}
```

#### 1.5.1.1. Via Invitation Link

This access allows users to access the box using an invitation link (key shares).

If another kind of access restriction exists, both the invitation link and the other access rules are required to access the box. Otherwise, only the invitation link is required.

```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (must be: invitation_link)",
        "value": "(string): the other_share_hash of the corresponding key share"
    },
    "referrer_id": null
}
```

#### 1.5.1.2. To a specific identifier

```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (must be: identifier)",
        "value": "(string): the identifier value"
    },
    "referrer_id": null
}
```

#### 1.5.1.3. To an entire email domain

```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (must be: email_domain)",
        "value": "(string): the email domain value (with no @)"
    },
    "referrer_id": null
}
```

### 1.5.2. Remove

The removal of an access is represented by this event shape:
```json
{
    "type": "access.rm",
    "referrer_id": "(string): the uuid of the corresponding access.add event that has been removed"
}
```

# 2. Privileges

## 2.1. Members

Members are users that have lastly joined the box.

## 2.2. Admins

Admins are considered as the users having the most advanced privileges on a box.

Today, only the creator of the box is considered as an admin.
