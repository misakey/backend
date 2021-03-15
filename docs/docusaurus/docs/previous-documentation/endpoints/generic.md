---
title: Generic
---

# Introduction

Here are some public endpoints.

They are not specific to the API but can serve multiple purposes: security, information, ...

## Get the version of the running instance

### Request

```bash
  GET https://auth.misakey.com/version
```

### success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "version": "v0.0.2"
}
```

- `version` (string): The API Version.

## Get the current CSRF Token

This endpoint allow to retrieve the CSRF Token that must be sent in the header
`X-CSRF-Token` with all HTTP verbs that can write data.

### Request

```bash
  GET https://auth.misakey.com/csrf
```

### success response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "csrf_token": "GMvLMlXGP9gaM7ZWbvBIQEwo4rZKVODL"
}
```

- `csrf_token` (string): The CSRF Token.

