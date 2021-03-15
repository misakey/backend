---
title: Coupons
---

## Introduction

**Coupons** are objects used to give insentive to the user to invite people to use the app.

A user can add a coupon in their identity to improve their identity (access to the app, more storage, ...)

They are linked to an identity.

## Attach a coupon to an identity

This route let a user attach a coupon to their identity.

#### Request

```bash
POST https://api.misakey.com/identities/:id/coupons
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks, delivered at the end of the auth flow

_Path Parameters:_
- `value` (string): the value of the coupon.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```
