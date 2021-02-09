+++
categories = ["Endpoints"]
date = "2020-12-09"
description = "Generic endpoints"
tags = ["generic", "api", "endpoints"]
title = "Generic"
+++

# 1. Introduction

Here are some public endpoints.

They are not specific to the API but can serve multiple purposes: security, information, ...

## 1.1. Get the version of the running instance

### 1.2.1. request

```bash
  GET https://auth.misakey.com/version
```

### 1.2.2. success response

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

## 1.1. Get the current CSRF Token

This endpoint allow to retrieve the CSRF Token that must be sent in the header
`X-CSRF-Token` with all HTTP verbs that can write data.

### 1.2.1. request

```bash
  GET https://auth.misakey.com/csrf
```

### 1.2.2. success response

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

