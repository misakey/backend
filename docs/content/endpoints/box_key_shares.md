---
title: Box - Key Shares (Key Splitting)
---

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

_Headers:_
- `Authorization` (opaque token) (ACR >= 2): no identity check, just a valid token is required.

_JSON Body:_
```json
{{% include "include/box-key-share.json" %}}
```

- `share` (string) (base64): one of the shares.
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

_Headers:_
- `Authorization` (opaque token) (ACR >= 1): no identity check, just a valid token is required.

_Path Parameters:_
- `other-share-hash` (string): the invitation hash of the key share.

_Code:_
```bash
    HTTP 200 OK
```

_JSON Body:_
```json
{{% include "include/box-key-share.json" %}}
```

- `share` (string) (base64): one of the shares.
- `other-share-hash` (string) (unpadded url-safe base64): a hash of the other share.
- `box_id` (string) (uuid): the box id linked to the key shares.
