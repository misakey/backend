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

## Deleting Crypto Actions

```bash
DELETE /accounts/e61d516d-716b-44de-b017-2307eb76fb8d/crypto/actions
```

```json
{
    "until_action_id": "e7a3c382-899f-4cf5-bbaa-dca3324ffca6"
}
```

Will delete all of the actions of this account
until action with the given ID,
and *including this action*.

Rules:

- account ID must match the one in the access token
  (you can only delete your own crypto actions)
- action with ID `until_action_id` must exist and be owned by the querier's account

