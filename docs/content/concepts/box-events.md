---
title: Box Events
---

# 1. Events

*Boxes* objects contain *events*.
An event represents the messages sent to the box
as well as changes made to the state of the box
(closing the box, and in the future other operations such as changing the title etc ...).

[Events are mainly used in the “boxes” endpoint described here.](/endpoints/boxes)

An event has the following fields:

- `id`, a UUID, (set by the server)
- `server_event_created_at` (format: `"2020-04-01T20:22:45.691Z"`)
  the time at which it was received by the server (set by the server)
- `sender`, an object representing the sender (set by the server)
  with the following shape:
  ```
  {{% include "include/event-sender.json" 2 %}}
  ```
- `type`, one of `create`, `msg.txt`, `msg.file`, `state.lifecycle`, `join`
- `content`, an object which shape depends on `type`

## 1.1. “Create” type event

**Purpose:**

A box starts with a `create` event
which is not posted by the client
but is instead inserted automatically by the backend
at the creation of the box.

- the `server_event_created_at` field is the time of creation of the box itself
- the `sender` of the event is the user who created the box
- for `content`, see below

**Content:**

An event of type `create` has the following content fields:

- `public_key` (string): the public key that must be used to encrypt messages for this box
- `title` (string): the title of the box

## 1.2. "Join" type event

**Purpose:**

An event of type `join` is automatically added to the box by the backend
when a user joins the box with an invitation link.

Only one `join` event can exist per couple box/sender.

**Content:**

Messages of type `join` have no content field.

## 1.3. "Message" type event

**Purpose:**

Messages are events with type `msg.txt` or `msg.file`. They allow the transfer of blob data or message text.

**Content:**

For all `msg.*`:

- `encrypted` (string) (base64): the encrypted message.
  Recall that files are sent separately from the message,
  so the size of a message event stays rather small.

For `msg.file` event, there is an additionnal information:

- `encrypted_file_id` (string) (uuid): a unique identifier used to store and download the file

## 1.4. "State" type event

**Purpose:**

State events are meant to update the state of a box: information that can be set to only one value and are overriden when updated.

Exhaustive list of possible state events:
- `state.lifecycle`: change the lifecycle of a box.

**Content:**

Type `state.lifecycle` have the following content fields:

- `state` (string): for which the only allowed value for now is `closed`.

# 2. Accesses

## 2.1. Admins

Admins are considered as the users having the most advanced access on a box.

Today, only the creator of the box is considered as a admin.

## 2.2. Participants

Participants are users that have joined the box via an invitation link.