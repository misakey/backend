+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Files Upload & Download endpoints"
tags = ["box", "files", "upload", "download", "api", "endpoints"]
title = "Files Upload & Download"
+++

# 1. Introduction

Boxes contain *events* of type `msg.file` that are linked to a blob data.
This documentation describe the way to upload/download blob data linked to the event.

[The shape and rules for box events are described here.](/concepts/box-events)

# 2. Files
## 2.3. Upload an encrypted file to a box

The upload of an encrypted file triggers the creation of a `msg.file` event then returns it.

### 2.3.1. request

```bash
POST https://api.misakey.com/boxes/:id/encrypted-files
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): just a valid access token
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Path Parameters:_
- `id` (uuid string): the box unique id the file is uploaded in.

_Multipart Form Data Body:_
- `encrypted_file` (binary): the encrypted data, this file size must be less than 8MB.
- `msg_encrypted_content` (string) (base64): encrypted content that will be store in the created `msg.file` event.

### 2.3.2. success response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body_:
```json
{
    "id": "cac1f19f-46eb-4be9-ba21-8346f1fd3838",
    "type": "msg.file",
    "content": {
        "encrypted": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
        "encrypted_file_id": "9b2c3cdc-d768-43ef-ba23-350b56b3d8ed"
    },
    "server_event_created_at": "2020-06-19T15:36:28.092359097Z",
    "sender": {{% include "include/event-identity.json" 4 %}}
}
```

### 2.3.3. notable error responses

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

## 2.4. Download an encrypted file

### 2.4.1. request

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

### 2.4.2. response

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
