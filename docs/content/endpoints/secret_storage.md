+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Accounts endpoints"
tags = ["sso", "accounts", "api", "endpoints"]
title = "SSO - Accounts"
+++

## 1. Introduction

The secret storage is a mechanism for the frontend to store the cryptographic secrets of an **account**. It replaces the previous *secret backup* mechanism.

These secrets are encrypted by the frontend with a key called *account root key*, sometimes abbreviated as *root key*.

The root key itself is stored in the secret storage, encrypted with the *password hash* (the output of Argon2 over the user's password).

## 2. Migrating an Account to Secret Storage

To migrate an account that is still using the secret backup mechanism.

### 2.1 request

*TODO*

### 2.2 response

*TODO*


## 3. Getting the Account Secret Storage

### 3.1. request

```bash
GET https://api.misakey.com/crypto/secret-storage
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as an identity id linked to the account.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

### 3.2. success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/secretStorageView-filled.json" %}}
```