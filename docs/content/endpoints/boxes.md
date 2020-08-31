---
title: Box - Boxes
---

# 1. Introduction

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.

[The shape and rules for box events are described here.](/concepts/box-events)


# 2. Boxes

## 2.1. Creating a Box

### 2.1.1. request

```bash
  POST https://api.misakey.com/boxes
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

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

### 2.1.2. response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```

The most important part is the `id` field
which must be used to interact with the box.

## 2.2. Getting a Box

### 2.2.1. request

```bash
  GET https://api.misakey.com/boxes/:id
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

### 2.2.2. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```

## 2.3. Getting public info about a box

This endpoint does not need any valid token, but it needs a valid `other_share_hash`
corresponding to the box to get.

### 2.3.1. request

```bash
  GET https://api.misakey.com/boxes/:id/public?other_share_hash=
```

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

_Query Parameters:_
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.

### 2.3.1. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "title": "<title of the box>"
}
```

## 2.4 Get the total count of boxes for the current user

### 2.4.1. request

This request allows to retrieval of information about accessible boxes list.

Today only the total count of boxes is returned as an response header.

```bash
  HEAD https://api.misakey.com/boxes
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): a valid token.

### 2.4.2. response

_Code:_
```bash
HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of boxes that the user can access.

## 2.5. Listing boxes

### 2.5.1. request

Users are able to list boxes they have an access to.

The returned list is automatically computed from the server according to the authorization
provided by the received bearer token.

```bash
GET https://api.misakey.com/boxes
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): a valid token.

_Query Parameters:_

Pagination ([more info](/concepts/pagination)) with default limit set to 10.

### 2.5.2. response

_Code:_
```bash
HTTP 200 OK
```

A list of event is returned.
```json
[
  {{% include "include/box-list.json" %}}
]
```

## 2.6 Delete a box

[Box admins](../../concepts/box-events/#21-admins) only are able to delete corresponding boxes.

A removed box sees its data completely removed from Misakey storage. This action is irreversible.

This action removes all data related to the box (events, key-shares...).

### 2.6.1 request

```bash
DELETE https://api.misakey.com/boxes/:id
```

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): the linked identity must be considered as a [box admin](../concepts/box-events/#21-admins).

_Path Parameters:_
- `id` (uuid string): the box id wished to be deleted.

_JSON Body:_
```json
{
  "user_confirmation": "delete"
}
```

- `user_confirmation` (string) (one of: _delete_, _supprimer_): the input the end-user has entered to confirm the deletion. The server will check if the value corresponds to some expected strings (cf one of).

### 2.6.2. response

_Code:_
```bash
HTTP 204 NO CONTENT
```

# 3. Events on boxes

## 3.1. Sending an Event to a Box

### 3.1.1. request

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

List of events that cannot be posted by clients:
- `create`: they are created by the backend during the creation of the box.
- `msg.file`: they are created by the backend during the upload of an encrypted file.

### 3.1.2. response

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

### 3.1.3. notable error reponses

**I - The box is closed:**

A box that has received a lifecycle closed event cannot received new events.

```bash
HTTP 409 CONFLICT
```

_JSON Body:_
```json
{
    "code": "conflict",
    "origin": "not_defined",
    "desc": "box is closed.",
    "details": {
        "lifecyle": "conflict"
    }
}
```

## 3.2. Getting Events in a Box

### 3.2.1. request

```bash
  GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

### 3.2.2. response

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

## 3.3. Upload an encrypted file to a box

The upload of an encrypted file triggers the creation of a `msg.file` event then returns it.

### 3.3.1. request

```bash
POST https://api.misakey.com/boxes/:id/encrypted-files
```
_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): just a valid access token.

_Path Parameters:_
- `id` (uuid string): the box unique id the file is uploaded in.

_Multipart Form Data Body:_
- `encrypted_file` (binary): the encrypted data, this file size must be less than 8MB.
- `msg_encrypted_content` (string) (base64): encrypted content that will be store in the created `msg.file` event.

### 3.3.2. success response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body_:
```
{
    "id": "cac1f19f-46eb-4be9-ba21-8346f1fd3838",
    "type": "msg.file",
    "content": {
        "encrypted": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
        "encrypted_file_id": "9b2c3cdc-d768-43ef-ba23-350b56b3d8ed"
    },
    "server_event_created_at": "2020-06-19T15:36:28.092359097Z",
    "sender": {
        "display_name": "898d05-test@misakey.com",
        "avatar_url": null,
        "identifier": {
            "value": "898d05-test@misakey.com",
            "kind": "email"
        }
    }
}
```

### 3.3.3. notable error responses

1 - The file is too big:

The file size is limited to 8MB.

_Code:_
```bash
HTTP 400 BAD REQUEST
```

```json
{
    "code": "bad_request",
    "origin": "not_defined",
    "desc": "size: the maximum file size is 8MB.",
    "details": {
        "size": "invalid"
    }
}
```

## 3.4. Download an encrypted file

### 3.4.1. request

```bash
GET https://api.misakey.com/encrypted-files/:id
```

_Path Parameters:_
- `id` (string) (uuid): the encrypted file id contained in the content of the `msg.file` event.

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): a valid access token.

### 3.4.2. response

_Code:_
```bash
HTTP 200 OK
```

_Headers:_
- `Content-Type: application/octet-stream`: the response is an octet stream.

_Octect Stream Body:_
```
  (the raw data of the encrypted file)
```

## 3.5. Reset the new events count for an identity

This endpoint allows to reset the new events count of a box for a given identity.

It is a kind of an acknowledgement and it must be used when the user want to mark the box as "read".

### 3.5.1. request

```bash
PUT https://api.misakey.com/boxes/:id/new-events-count/ack
```

_Path Parameter:_
- `id` (string) (uuid): the box id to mark as "read"

_JSON Body:_
```json
    {
        "identity_id": "e2e49259-f840-4991-a9f7-97c5f267bd18"
    }
```

where `identity_id` is the identity of the requester who wants to acknowledge.

_Headers:_
- :key: `Authorization` (opaque token) (ACR >= 1): a valid access token corresponding to the identity of the body

### 3.5.2. success response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## 3.6. Deleting a Message Event

A message (text or file) can be deleted by its author or by the box admin.

The event is still present in the box but the encrypted content is removed
and replaced by who deleted it and when.

Note that, contrary to what the endpoint may suggest,
no new event is added to the box with this call.

### 3.6.1. Request

```bash
POST /boxes/:id/events
```

```json
{
  "type": "msg.delete",
  "content": {
    "event_id": "f17169e0-61d8-4211-bb9f-bac29fe46d2d"
  }
}
```

Where `event_id` is the ID of the event to delete.

- The sender's account must be the one that sent the event to delete,
  or the sender must be the box creator.
- the message must not be already deleted
- the box must be not be closed

### 3.6.2. Success Response

```bash
HTTP 201 Created

{
    "type": "msg.text",
    "content": {
        "deleted": {
            "at_time": "2020-08-25T15:58:28.259977262Z",
            "by_identifier_id": "1bcdd455-f42c-43ce-8109-6a8c1eb9d4c0"
        }
    },
    "id": "c53d6334-a853-4f23-a143-d15a5a60340c",
    "server_event_created_at": "2020-08-25T15:58:28.246201Z",
    "sender": {
        "display_name": "38a241-Test",
        "avatar_url": null,
        "identifier": {
            "value": "38a241-test@misakey.com",
            "kind": "email"
        }
    }
}
```

### 3.6.3 Notable Message Errors

Already deleted:

```bash
HTTP 410 Gone
```

```json
{
    "code": "gone",
    "origin": "not_defined",
    "desc": "event is already deleted",
    "details": {}
}
```

## 3.7. Editing a Message

Users can edit their own message

### 3.7.1 Request

```bash
POST /boxes/a7f7a0c5-3061-41e0-bf83-123b8e5fd3d0/events
```

```json
{
    "type": "msg.edit",
    "content": {
        "event_id": "7410feae-637e-40a8-ab59-badeaf479c63",
        "new_encrypted": "EditedXXB64dcc9PhJTeyUS2K04zeHKLMW8fviUkmyBjWdGvwwo=",
        "new_public_key": "EditedXXa75RO1FzZpskiKHAggyB7YNJoz4R24dnMFvHfMzu4wQ="
    }
}
```

Where `event_id` is the ID of the event to edit.

- The sender's account must be the one that sent the event to edit.
- the message must not be already deleted
- the box must be not be closed

### 3.7.2 Success Response

```bash
HTTP 201 Created
```

```json
{
    "type": "msg.text",
    "content": {
        "encrypted": "EditedXXB64dcc9PhJTeyUS2K04zeHKLMW8fviUkmyBjWdGvwwo=",
        "public_key": "EditedXXa75RO1FzZpskiKHAggyB7YNJoz4R24dnMFvHfMzu4wQ=",
        "last_edited_at": "2020-08-25T15:58:27.79441082Z"
    },
    "id": "7410feae-637e-40a8-ab59-badeaf479c63",
    "server_event_created_at": "2020-08-25T15:58:27.751261Z",
    "sender": {
        "display_name": "326b00-Test",
        "avatar_url": null,
        "identifier": {
            "value": "326b00-test@misakey.com",
            "kind": "email"
        }
    }
}
```

### 3.7.3 Notable Error Responses

- type cannot be edited: `HTTP 401 Unauthorized`
- event has been deleted: `HTTP 410 Gone`