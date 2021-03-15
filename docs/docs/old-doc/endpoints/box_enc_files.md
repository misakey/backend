---
title: Files Upload & Download
---

# Introduction

Boxes contain *events* of type `msg.file` that are linked to a blob data.
This documentation describe the way to upload/download blob data linked to the event.

[The shape and rules for box events are described here.](/old-doc/concepts/box-events.md)

# Files

## Upload an encrypted file to a box

[Moved here](https://docs.misakey.com/docs/references/boxes#send-data-as-a-file-to-a-box).

## Download an encrypted file
#### Request

```bash
GET https://api.misakey.com/encrypted-files/:id
```

_Path Parameters:_
- `id` (string) (uuid): the encrypted file id contained in the content of the `msg.file` event.

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): just a valid access token
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

#### Response

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
