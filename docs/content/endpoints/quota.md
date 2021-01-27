+++
categories = ["Endpoints"]
date = "2020-09-29"
description = "Quota endpoints"
tags = ["quota", "storage", "used space", "vault", "box", "api", "endpoints"]
title = "Storage quota"
+++

## 1. Introduction

### 1.1. Concept

Storage quota are used to compute already used storage and max autorized storage for a user.

[The shape and rules for storage objects are described here.](/concepts/quota)

## 2. Get storage quota

This route is used to retrieve all the `storage_quotum` object associated to an identity

### 2.1. request

```bash
GET https://api.misakey.com/box-users/:id/storage-quota
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity id.

### 2.2. success response

_Code:_

```bash
HTTP 200 OK
```

```json
[
    {
        "id": "e989f6fd-7b01-4e1b-b05b-5ad41bc71af3",
        "origin": "base",
        "identity_id": "b88d88e2-b99c-40d6-bcaa-6efc91d00bfd",
        "value": 104857600
    }
]
```

## 3. Get current user storage use

This route is used to retrieve all the `box_used_space` object associated to an identity

### 3.1. request

```bash
POST https://api.misakey.com/box-used-spaces?&identity_id=
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Query Parameters:_
- `identity_id` (uuid string): the identity id.

### 3.2. success response

_Code:_

```bash
HTTP 200 OK
```

```json
[
    {
        "box_id": "e989f6fd-7b01-4e1b-b05b-5ad41bc71af3",
        "value": 16475271
    }   
]
```

## 4. Get vault used space

This route is used to retrieve the vault used space linked to an identity.

### 4.1. request

```bash
POST https://api.misakey.com/box-users/:id/vault-used-space
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity id.

### 4.2. success response

_Code:_

```bash
HTTP 200 OK
```

```json
{
    "value": 5271
}
```
