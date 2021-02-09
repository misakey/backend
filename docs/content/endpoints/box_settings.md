+++
categories = ["Endpoints"]
date = "2020-10-23"
description = "Box settings endpoints"
tags = ["box", "settings", "api", "endpoints"]
title = "Box Settings"
+++

## 1. Introduction

The box settings are unique per user settings for a given box.

We don’t manage it with events because it is specific to each user and does not describe
the box state.

It is stored in a different place and is not returned with the actual box.

## 2. Updating a Box Setting

### 2.1. request

```bash
  PUT https://api.misakey.com/box-users/:id/boxes/:bid/settings
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): identity must be the same than the requested identity.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (string): the identity id
- `bid` (string): the box id

_JSON Body:_
```json
{
    "muted": true|false
}
```

- `muted` (bool): is the user notified on a box update;

### 2.2. response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## 2. Getting a Box Setting

### 3.1. request

```bash
  GET https://api.misakey.com/box-users/:id/boxes/:bid/settings
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): identity must be the same than the requested identity.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (string): the identity id
- `bid` (string): the box id

### 3.2. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "identity_id": "41e213d1-7d85-4b08-b913-678da2653021",
    "box_id": "89e213d1-7d85-4b08-b913-678da2653846",
    "muted": true|false
}
```

- `identity_id` (string) (uuid): the identity id.
- `box_id` (string) (uuid): the box id.
- `muted` (bool):is the user notified on a box update?
