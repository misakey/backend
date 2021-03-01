+++
categories = ["Endpoints"]
date = "2020-12-31"
description = "Datatag Endpoints"
tags = ["sso", "datatag", "api", "endpoints"]
title = "Datatag"
+++

## 1. Introduction

### 1.1. Concept

Datatags are used to identify data shipped through Misakey channels.

They are simple strings defining a data type and they depend on a single organization.

## 2. Datatags

### 2.1. Create datatag

#### 2.1.1. request

```bash
POST https://api.misakey.com/organizations/:id/datatags
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the organization admin.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path parameter:_
- `id`: an uuid to identify the organization

_JSON Body:_
```json
{
    "name": "<string>"
}
```

- `name`: the datatag name

#### 2.1.2. success response

_Code:_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{
    "id": "<uuid string>",
    "organization_id": "<uuid string>",
    "name": "<string>",
    "created_at": "<date string>"
}
```

- `id`: an uuid to identify the datatag
- `organization_id`: an uuid to identify the organization
- `name`: the datatag name
- `created_at`: date of creation

### 2.2. Edit datatag

#### 2.2.1. request

```bash
PATCH https://api.misakey.com/organizations/:id/datatags/:did
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the organization admin.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameter:_
- `id`: an uuid to identify the organization
- `did`: an uuid to identify the datatag

_JSON Body:_
```json
{
    "name": "<string>"
}
```

- `name`: the datatag name

#### 2.2.2. success response

_Code:_
```bash
HTTP 204 NO CONTENXT
```
### 2.3. List datatags

#### 2.3.1. request

```bash
GET https://api.misakey.com/organizations/:id/datatags
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as the organization admin.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameter:_
- `id`: an uuid to identify the organization

#### 2.3.2. success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
  {
    "id": "<uuid string>",
    "organization_id": "<uuid string>",
    "name": "<string>",
    "created_at": "<date string>"
  }
]
```

- `id`: an uuid to identify the datatag
- `organization_id`: an uuid to identify the organization
- `name`: the datatag name
- `created_at`: date of creation

