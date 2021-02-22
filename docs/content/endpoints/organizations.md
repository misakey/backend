+++
categories = ["Endpoints"]
date = "2020-08-01"
description = "Organization endpoints"
tags = ["org", "api", "endpoints", "orga", "organization", "organizations"]
title = "Organizations"
+++

# 1. Introduction

An organization is created by an end-user. Inside it, they can create boxes and administrate them. These boxes are then owned by the organization.

There is always at least one organization within the system, corresponding to what is called the "self org".
The "self org" corresponds to the Open ID Provider client and represents the instance of the system running.

While end-users create boxes in their personal space, it is linked to this self-organization which represent then the personal space for all the users on this instance.
Self organization has no administrators, the data linked to it belongs to the end-users that have created it.

# 2. Organizations

## 2.1. Creating an Organization

### 2.1.1. request

```bash
  POST https://api.misakey.com/organizations
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): no identity check, just a valid token is required.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_JSON Body:_
```json
    {
      "name": "The Privacy-Esteeming Org",
    }
```

- `name` (string) (max length: 255).

### 2.1.2. response

_
```bash
HTTP 201 Created
```

_JSON Body:_
```json
{{% include "include/org.json" %}}
```

{{% include "include/org.md"  %}}

## 2.2. Listing organizations for the current identity

### 2.2.1. request

```bash
  GET https://api.misakey.com/identities/:id/organizations
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the box id wished to be retrieved.

### 2.2.2. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
  {{% include "include/org.json" 2 %}}
]
```

{{% include "include/org.md"  %}}


## 2.3. Getting public info about an organization

This endpoint does not need any valid token.

### 2.3.1. request

```bash
  GET https://api.misakey.com/organizations/:id/public
```

_Path Parameters:_
- `id` (uuid string): the organization id wished to be retrieved.

### 2.3.2. response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "id": "<(uuid string): the organization id>",
    "name": "<name of the organization>",
    "logoUrl": "<logo of the organization>",
}
```
