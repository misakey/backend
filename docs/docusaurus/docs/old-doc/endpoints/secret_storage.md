---
title: Secret Storage
---

## Introduction

The secret storage is a mechanism for the frontend to store the cryptographic secrets of an **account**. It replaces the previous *secret backup* mechanism.

These secrets are encrypted by the frontend with a key called *account root key*, sometimes abbreviated as *root key*.

The root key itself is stored in the secret storage, encrypted with the *password hash* (the output of Argon2 over the user's password).

## Migrating an Account to Secret Storage

To migrate an account that is still using the secret backup mechanism.

#### Request

*TODO*

#### Response

*TODO*


## Getting the Account Secret Storage

#### Request

```bash
GET https://api.misakey.com/crypto/secret-storage
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): `mid` claim as an identity id linked to the account.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "account_root_key": {
        "key_hash": "V2Wrv7LjW_Ct_LPTuDcHSg",
        "encrypted_key": "dsxTchS2R75QayOvAYFe4A"
    },
    "vault_key": {
        "key_hash": "JoUZHeHJ6zaz6vRw_Padgg",
        "encrypted_key": "RDTEMIlE4JlMAofbaF_VpQ"
    },
    "asym_keys": {
        "AbKozlB19lF8mGp0NURW0A": {
            "encrypted_secret_key": "DG5x6eRi9W56_BJpst5dgZCmu0O-s2gmkY_CPFkutF8"
        },
        "TTeZCTxUVYhQFHXGnpjcYA": {
            "encrypted_secret_key": "Xqyhq3QXcsiFoOwqUDr8GQ"
        },
        "YobcbIi45V68XFOyU_q4nQ": {
            "encrypted_secret_key": "Tl53fZahxhijbuWH3qWkqOSDodt5UhlyNRVSyAegIpY"
        }
    },
    "box_key_shares": {
        "a6ca4e2b-d929-49eb-8bdd-456d46f5b098": {
            "id": "d722a958-20b0-4ece-b5b2-6312a71415f8",
            "invitation_share_hash": "3A84CQUy0I4Opv-vlDuiHg",
            "encrypted_invitation_share": "-EiyX8wGr23gVc55aA3llycRhJ1LAWYCkjdwLS-AVHE",
            "created_at": "2021-02-11T08:57:13.741859Z",
            "updated_at": "2021-02-11T08:57:13.741859Z"
        }
    }
}
```