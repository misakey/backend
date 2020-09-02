---
title: SSO - Coupons
---

## 1. Introduction

### 1.1. Concept

**Coupons** are objects used to give insentive to the user to invite people to use the app.

A user can add a coupon in their identity to improve their identity (access to the app, more storage, ...)

They are linked to an identity.

## 2. Attach a coupon to an identity

This route let a user attach a coupon to their identity.

### 2.1. request

```bash
POST https://api.misakey.com/identities/:id/coupons
```
_Headers:_
- `Authorization` (opaque token) (ACR >= 1): `mid` claim as the identity id.

_Path Parameters:_
- `value` (string): the value of the coupon.

### 2.2. success response

_Code:_
```bash
HTTP 204 NO CONTENT
```
