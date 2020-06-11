---
title: Box Events
---

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
- `type`, one of `create`, `msg.txt`, `msg.file`, `state.lifecycle`
- `content`, an object which shape depends on `type`

## “Create”-type event

An event of type `create` is automatically added to the box by the backend
when the box is created:

- the `server_event_created_at` field is the time of creation of the box itself
- the `sender` of the event is the user who created the box
- for `content`, see below

## Content of Each Type of Event

A box starts with a `create` event
which is not posted by the client
but is instead inserted automatically by the backend
at the creation of the box.
An event of type `create` has the following content fields:

- `public_key`: the public key that must be used to encrypt messages for this box
- `title`: the title of the box

Messages are events with type `msg.txt` or `msg.file` with the following content fields:

- `encrypted` is the encrypted message (encoded with base64).
  Recall that files are sent separately from the message,
  so the size of a message event stays rather small.

Messages of type `state.lifecycle` have the following content fields:

- `state` for which the only allowed value for now is `closed`.