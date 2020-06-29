---
title: Box - Boxes
---

Boxes contain *events* that have a *type*.
In practice, most events will be of type `msg.text` or `msg.file`,
corresponding to the sending of messages (with either text or files in it) to the box.
There are however a few other events,
most of them describing a change of the *state* of the box.
[The shape and rules for box events are described here.](/concepts/box-events)


## Creating a Box

### Request

```bash
    POST https://api.misakey.com/boxes
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

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

### Response

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

## Getting a Box

### Request

```bash
    GET https://api.misakey.com/boxes/:id
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

### Response

_Code:_
```bash
    HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/box.json" %}}
```

## Get the total count of boxes for the current user

### Request

This request allows to retrieval of information about accessible boxes list.

Today only the total count of boxes is returned as an response header.

```bash
  HEAD https://api.misakey.com/boxes
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): a valid token.

### Response

_Code:_
```bash
  HTTP 204 NO CONTENT
```

_Headers:_
- `X-Total-Count` (integer): the total count of boxes that the user can access.

## Listing boxes

### Request

Users are able to list boxes they have an access to.

The returned list is automatically computed from the server according to the authorization
provided by the received bearer token.

```bash
  GET https://api.misakey.com/boxes
```

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): a valid token.

_Query Parameters:_

Pagination ([more info](/concepts/pagination)) with default limit set to 10.

### Response

_Code:_
```bash
  HTTP 200 OK
```

A list of event is returned.
```json
[
  {{% include "include/box.json" %}}
]
```

## Sending an Event to a Box

### Request

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

### Response

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

### Notable Error Reponses

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

## Getting Events in a Box

### Request

```bash
    GET https://api.misakey.com/boxes/74ee16b5-89be-44f7-bcdd-117f496a90a7/events
```

### Response

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

## Upload an encrypted file to a box

The upload of an encrypted file triggers the creation of a `msg.file` event then returns it.

### Request

```bash
  POST https://api.misakey.com.local/boxes/:id/encrypted-files
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): just a valid access token.

_Path Parameters:_
- `id` (uuid string): the box unique id the file is uploaded in.

_Multipart Form Data Body:_
- `encrypted_file` (binary): the encrypted data, this file size must be less than 8MB.
- `msg_encrypted_content` (string) (base64): encrypted content that will be store in the created `msg.file` event.

### Success Response

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

### Notable Error Reponses

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

## Download an encrypted file

### Request

```bash
  GET https://www.api.misakey.com/boxes/:bid/encrypted-files/:eid
```

_Path Parameters:_
- `bid` (string) (uuid): the box id where the file has been initially uploaded.
- `eid` (string) (uuid): the encrypted file id contained in the content of the `msg.file` event.

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): a valid access token.

### Response

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