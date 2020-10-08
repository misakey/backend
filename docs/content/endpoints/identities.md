+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Identities endpoints"
tags = ["sso", "identities", "api", "endpoints"]
title = "SSO - Identities"
+++

## 1. Introduction

### 1.1. Concept

**Identity** is the core entity used for authorization and authentication at Misakey.

They are auto-generated: the system consider a new wild identity appears when an
unknown identifier is entered during a login flow.

They refer then to an identifier (email, phone number...) and they are optionally attached to [an account](../accounts/).

End-users can claim an identity by creating an account on it or link an existing account.

They can be considered both as profiles and concrete identities. Users have many identities:
- as citizens: having both Korean and Russian nationalities...
- as internet fellows: having a Travian account and a Mastodon account...

### 1.2. Resource Access

Resources access is based on the identity the end-user is logged in.

The end-user can select the identity they want to be logged in during the auth flow,
or in the interface at any moment.

There is no need to re-authenticate switching identities unless the security level
of an identity is higher than the current one the end-user is logged in.

## 2. Require an authable identity for a given identifier

Described in [the auth flow section](../auth_flow/#5-require-an-authable-identity-for-a-given-identifier).

## 3. Create an account on an identity

Described [in the accounts section](../accounts/#2-create-an-account-on-an-identity)

## 4. Get an identity

This route allows the retrieval of the information related to an identity.

### 4.1. request

```bash
GET https://api.misakey.com/identities/:id
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `mid` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

### 4.2. success response

_Code:_
```bash
HTTP 200 CREATED
```

_JSON Body:_
```json
  {
    "id": "89a27dec-c0bb-40ed-bfc8-dc74a1b99dc9",
    "account_id": null,
    "has_account": false,
    "is_authable": true,
    "display_name": "iamagreat@dpo.com",
    "notifications": "minimal",
    "avatar_url": null,
    "confirmed": true,
    "identifier_id": "abc9dd47-0c4e-4ef3-ae27-0c21ea6d7450",
    "identifier": {
      "id": "e5d889de-6be1-4201-bb7e-0772fbbf41e2",
      "value": "iamagreat@dpo.com",
      "kind": "email"
    }
  }
```

- `id` (uuid string): the unique identity id.
- `account_id` (uuid string) (nullable): the linked account unique id, always null if the end-user is connected with ACR 1.
- `has_account` (boolean): tell either the identity is linked or not to an account.
- `is_authable` (uuid string): either the identity can be used in a login flow.
- `display_name` (uuid string): the name to display to represent the identity.
- `notifications` (uuid string): the frequency of notifications for this identity.
- `avatar_url` (uuid string) (nullable): the web-address of the avatar's file content.
- `confirmed` (uuid string): either the identity has been guaranted to be owned by the end-user.
- `identifier_id` (uuid string): the linked identifier unique id.
- `identifier` (json object): the linked identifier object, nested.
  - `id` (uuid string): the unique identifier id.
  - `kind` (string) (oneof: _email_): the kind of the identifier.
  - `value` (string): the value of the identifier.

## 5. Update an identity

For the moment, only the Display Name and Notifications can be updated.

The request must be authenticated with a token corresponding to the updated identity.

### 5.1. request

```bash
PATCH https://api.misakey.com/identities/:id
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `mid` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Body Parameters:_
- `display_name` (string): the identity display name.
- `notifications` (string): notification setting. Must be one of `minimal`, `moderate`, `frequent`.

### 5.2. success response

_Code:_
```bash
HTTP 204 No Content
```

## 6. Upload an avatar

The request must be authenticated with a token corresponding to the identity on which the avatar is uploaded.

### 6.1. request

```bash
PUT https://api.misakey.com/identities/:id/avatar
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `mid` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Body Parameters (multipart/form\_data):_
- `avatar` (object): the avatar file.

### 6.2. success response

_Code:_
```bash
HTTP 204 No Content
```

## 7. Delete an avatar

The request must be authenticated with a token corresponding to the identity on which the avatar is deleted.

If no avatar is set on the identity, the request will return a `409 CONFLICT`.

### 7.1. request

```bash
DELETE https://api.misakey.com/identities/:id/avatar
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `mid` claim as the identity id.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

### 7.2. success response

_Code:_
```bash
HTTP 204 No Content
```
