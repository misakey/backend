+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Crypto Actions endpoints"
tags = ["sso", "crypto", "actions", "api", "endpoints"]
title = "SSO - Crypto Actions"
+++

Crypto actions are encrypted messages sent automatically from account to account,
often to give the recipient account a new cryptographic secret (key share, key...).

The frontend regularly lists its crypto actions,
processes them, and deletes the ones that have been processed.

Right now the only type of crypto action is `invitation`,
in which case the `encrypted` part decrypts to
the user key share of the box identified by `box_id`
(this is the same data one receives through an invitation link).

## Listing One's Crypto Actions

```bash
GET /accounts/e61d516d-716b-44de-b017-2307eb76fb8d/crypto/actions
```

Rules:

- account ID must match the one in the access token
  (you can only list your own crypto actions)

Response:

```bash
HTTP 200 OK
```
```json
[
    {
        "id": "e7a3c382-899f-4cf5-bbaa-dca3324ffca6",
        "type": "invitation",
        "box_id": "(uuid or null)",
        "encryption_public_key": "(URL-safe unpadded base64)",
        "encrypted": "(encrypted data)",
        "created_at": "2020-09-04T09:13:34.508851Z"
    }
]
```

## Getting a Specific Crypto Action

```bash
GET /accounts/e61d516d-716b-44de-b017-2307eb76fb8d/crypto/actions/e7a3c382-899f-4cf5-bbaa-dca3324ffca6
```

Rules:

- account ID must match the one in the access token
  (you can only list your own crypto actions)

Response:

```bash
HTTP 200 OK
```
```json
{
    "id": "e7a3c382-899f-4cf5-bbaa-dca3324ffca6",
    "type": "invitation",
    "box_id": "(uuid or null)",
    "encryption_public_key": "(URL-safe unpadded base64)",
    "encrypted": "(encrypted data)",
    "created_at": "2020-09-04T09:13:34.508851Z"
}
```

## Deleting a Crypto Action

```bash
DELETE /accounts/e61d516d-716b-44de-b017-2307eb76fb8d/crypto/actions/e7a3c382-899f-4cf5-bbaa-dca3324ffca6
```

Rules:

- account ID must match the one in the access token
  (you can only delete your own crypto actions)
- action with this ID must exist and be owned by the querier's account

Side effects:

- notifications pointing to this crypto action
  (e.g. auto invitations) will be marked as “used”.

Response if success: `HTTP 204 No Content`
