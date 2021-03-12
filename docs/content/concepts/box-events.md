+++
categories = ["Concepts"]
date = "2020-09-11"
description = "Box Events"
tags = ["concepts", "box", "events"]
title = "Box Events"
+++


# 1. Identity View

Identity view is an object representing an actor within event worlds.

The identity contains complementary information and has some removed considering user privacy and wishes (profile configuration).

It can sometimes be bound to the key `sender`, `kicked`, `deletor`...

:warning: Consumer can rely on the identity id as it is the only information about the end-user always presented.

Here is the description of this view:

```json
{{% include "include/event-identity.json" %}}
```

{{% include "include/event-identity.md"  %}}

# 2. Events

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
  "box_id": "74ee16b5-89be-44f7-bcdd-117f496a90a7",
  "sender": {{% include "include/event-identity.json" 2 %}},
  "type": "(string) (one of: create, msg.txt, msg.file, , state.key_share, state.access_mode, member.join, member.leave): the type of the event",
  "content": "(json object) (nullable): its shape depends on the type of event - see definitions below",
  "referrer_id": "(string) (uuid) (nullable): the uuid of a potential referrer event"
}
```

## 2.1. `Create` type event

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
    "owner_org_id": "(uuid string) the organization id",
    "datatag_id": "(uuid string) (optional) the datatag id",
    "subject_identity_id": "(uuid string) (optional) the subject identity id",
  },
  "referrer_id": null
}
```

## 2.2. `Member` type events

### 2.2.1. Join

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

### 2.2.2. Leave

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

### 2.2.3. Kick

An event of type `member.kick` is automatically added to the box by the backend
when an `access.rm` removes the last access rules allowing some members to access the box.

If any other rules allows the members to access the box, the event is not triggered.

The event refers the previous `member.join` event of the user, the `sender_id` is set to the kicked member id.

Messages of type `member.kick` have no content.

```json
{
  "type": "member.kick",
  "referrer_id": "(string) (uuid): member.join id",
  "sender_id": "(string) (uuid): the kicked identity",
  "content": {
      "kicker_id": "(string) (uuid): the person who has triggered the kick event"
  },
}
```

Once the event has been created for a user, the user cannot get/list the box anymore.

On read, the `kicker_id` is transformed into a `kicker` field containing sender information. This `kicker` attribute is nullable.

## 2.3. `Message` type events

Messages allow the transfer of encrypted message text or blob data.

### 2.3.1. Message Text
Messages of type `msg.txt` allow the transfer of message text.

```json
{
  "type": "msg.txt",
  "content": {
    "encrypted": "(string) (unpadded URL-safe base64): the encrypted message. Recall that files are sent separately from the message, so the size of a message event stays rather small.",
    "deleted": { // nullable, indicates the message have been removed
      "at_time": "indicates the deletion time of the message",
      "by_identity": "indicates who has deleted the message",
    },
    "last_edited_at": "(RFC3339 time): indicates that the message was edited, and when",
    "referrer_id": null
}
```

### 2.3.2. Message File

Messages of type `msg.file` allow the transfer of blob data.

```json
{
  "type": "msg.file",
  "content": {
    "is_saved": "(bool): is the file in user My Documents",
    "encrypted": "(string) (unpadded URL-safe base64): information about file encryption.",
    "encrypted_file_id": "(string) (uuid format): a unique uuid used to store and download the file",
    "deleted": { // nullable, indicates the message have been removed
      "at_time": "indicates the deletion time of the message",
      "by_identity": "indicates who has deleted the message",
    },
    "referrer_id": null
}
```

### 2.3.3. Deleting a Message Event

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

### 2.3.4. Editing a Message

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

## 2.4. `State` type events

State events are meant to update the state of a box: information that can be set to only one value and are overriden when updated.

### 2.4.1. `State Access Mode`

The `state.access_mode` event changes the access mode of a box: public or limited.

Anyone can join a box in public mode.
Only people matching access rules can join a box in a limited mode.

```json
{
  "type": "state.access_mode",
  "content": {
    "value": "(string) (one of: public, limited): the new access mode value of the box"
  },
  "referrer_id": null
}
```

### 2.4.2. `State Key Share`

`state.key_share` events change the key share of a box.

```json
{
    "type": "state.key_share",
    "extra": {
        "misakey_share": "BBVZBhrLtb0DsdYtul7s1g==",
        "other_share_hash": "h1vUkzYPYwaRgH03-4L7-g",
        "encrypted_invitation_key_share": "cp3nvY+OtRtetFGN0Yuxw3Cra6OjbWzO1ptOWP9hcWo="
    }
}
```

Remarks:
- `other_share_hash` is encoded in *unpadded URL-safe base64*.
- this event does not have a `content` field

Side effects:
- the key share of the box will be changed (and all previous ones for this box are deleted)
- every ACR2 member of the box will receive a cryptoaction
  with type `set_box_key_share`
  and encrypted content the value of `extra.encrypted_invitation_key_share`.

For more information on box key shares, see [here](/endpoints/box_key_shares).

## 2.5. `Access` type events

Access events are specific rules defined by the admins and allowing considering their logic who can access the box.

### 2.5.1. Add

The add of an access is represented by this event shape
```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (one of : identifier or email_domain): the type of restriction the access bears",
        "value": "(string): a value describing the restriction",
        "auto_invite": "(optional) (boolean)",
    },
    "referrer_id": null
}
```

Where `auto_invite` is only required if `restriction_type` is `identifier` under some circumstances
(see Section “to a specific identifier value”).

Note that there is no `content.value` field (it has been deprecated).

#### 2.5.1.1. To a specific identifier value

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

Or, in order to automatically invite the identifier
(through [crypto actions](/concepts/crypto-actions)):

```json
{
    "type": "access.add",
    "content": {
        "restriction_type": "(string) (must be: identifier)",
        "value": "(string): the identifier value",
        "auto_invite": true
    },
    "extra": {
      "a6cBxJMq8": "(encrypted)",
      "b7x94c1wG": "(encrypted)"
    }
}
```

Where `extra` is a mapping from each identity public key of the identifier being added
to the encrypted crypto action to send to the corresponding identity.
This will create a crypto action with type `invitation`
and a notification with type `box.auto_invite`.

#### 2.5.1.2. To an entire email domain

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

### 2.5.2. Remove

The removal of an access is represented by this event shape:
```json
{
    "type": "access.rm",
    "referrer_id": "(string): the uuid of the corresponding access.add event that has been removed"
}
```

# 3. Privileges

## 3.1. Members

Members are users that have lastly joined the box.

## 3.2. Admins

Admins are considered as the users having the most advanced privileges on a box.

Today, only the creator of the box is considered as an admin.
