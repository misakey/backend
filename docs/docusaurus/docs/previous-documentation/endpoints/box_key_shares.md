---
title: Box Key Shares
---

## Introduction

Key Splitting consists in splitting a secret key in several (currently, always two) *key shares*.
One share alone is completely useless, but by combining two shares of a key one can recover the secret key.

We use this technique for invitation links in boxes:
instead of sending the box secret key to the guest, we send a share of the key.
This makes the link less sensitive: it mitigates the security consequences of an invitation link ending up in malicious hands.

Of course the guest will need the other share to be able to access the box.
This second share is sent to Misakey's backend.
Because Misakey only sees this share and not the one in the invitation link,
end-to-end encryption is not compromised.

This is how key splitting happens:
when the creator of a box (or any user having the box secret key and who we allow to create invitations)
requires an invitation link, her frontend creates two shares of the box secret key.
One is sent (via `HTTP POST`) to Misakey, the other is encoded in the invitation link.

When the frontend of the guest receives the invitation link,
first it must authenticate,
and then it will ask the backend for the other key share.

A key share has another attribute than its value,
it has an `invitation_hash` which is used for the guest frontend to identify which share it wants to retrieve.
Technically speaking, the invitation hash is the SHA-512 hash of the share sent in the invitation.

## Creating a Box Key Share

### Request

```bash
  POST https://api.misakey.com/box-key-shares
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): no identity check, just a valid token is required.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_JSON Body:_
```json
{{% include "include/box-key-share.json" %}}
```

- `share` (string) (unpadded url-safe base64): one of the shares.
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.
- `box_id` (string) (uuid): the box id linked to the key shares.

### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{{% include "include/box-key-share.json" %}}
```

## Getting a Box Key Share

### Request

```bash
  GET https://api.misakey.com/box-key-shares/:other-share-hash
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): no identity check, just a valid token is required.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `other-share-hash` (string): the invitation hash of the key share.

### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/box-key-share.json" %}}
```

- `share` (string) (unpadded url-safe base64): the misakey share.
- `other-share-hash` (string) (unpadded url-safe base64): a hash of the other share (invitation share).
- `box_id` (string) (uuid): the box id linked to the key shares.

Note that box key shares now also have an *encrypted invitation key share*
(see next section)
that does not appear in this endpoint.

## Getting an Encrypted Box Key Share

Actually, this endpoint gives you the *encrypted invitation key share* part
of the *box key share* object.

Request:
```bash
GET https://api.misakey.com/box-key-shares/encrypted-invitation-key-share?box_id=74ee16b5-89be-44f7-bcdd-117f496a90a7
```

Access control:
- querier must be authenticated with ACR ≥ 2

Response:
```json
"cGYMzgIO9rc03WoSLAyoiQdLu7he5VbMRImLhRPmwTQ"
```

Not that the result is not a JSON literal object but it is still valid JSON
(a JSON string is a valid JSON object).
The encoding is unpadded URL-safe base64.
