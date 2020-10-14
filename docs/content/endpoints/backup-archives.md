+++
categories = ["Endpoints"]
date = "2020-09-11"
description = "Backup Archives endpoints"
tags = ["sso", "backup", "archives", "api", "endpoints"]
title = "SSO - Backup Archives"
+++


A new backup archive is created whenever a user successfully performs a “forgotten password” procedure.
The purpose of a backup archive is to keep a copy of the user's secret backup
(see [endpoints related to account](/endpoints/accounts))
at the time the password reset happenned,
since reseting a user's password implies to overwrite her backup with an empty new one.

After a backup archive has been created,
the user can attempt to decrypt it
(by remembering the password that was lost, or providing the corresponding backup key).

The user can ask the backend to delete a backup archive,
typically if she fears the corresponding password or backup key might have been exposed.

The frontend will also request the deletion of an archive after it has been successfully recovered.

# Listing Backup Archives

```bash
GET https://api.misakey.com/backup-archives
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity must be linked to an account
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

Only the backup archives related to the querier's account will be returned.

## Response

```bash
HTTP 200 OK
```

```json
[
    {
        "id": "99e0fad2-1191-44b7-9f78-4f10a8594c99",
        "account_id": "1dc726b3-895f-4d6f-a285-d53ae0dd1948",
        "created_at": "2020-08-05T16:02:22.613643Z",
        "recovered_at": null,
        "deleted_at": null
    },
    {
        "id": "51b2fc64-3009-496f-95c2-315705c6fbd1",
        "account_id": "1dc726b3-895f-4d6f-a285-d53ae0dd1948",
        "created_at": "2020-08-05T16:02:22.104236Z",
        "recovered_at": null,
        "deleted_at": null
    },
    {
        "id": "8ca33119-85cc-4638-af46-8a6e4c4dabd8",
        "account_id": "1dc726b3-895f-4d6f-a285-d53ae0dd1948",
        "created_at": "2020-08-05T16:02:21.27032Z",
        "recovered_at": null,
        "deleted_at": null
    }
]
```

# Getting the Data of a Backup Archive

```bash
GET https://api.misakey.com.local/backup-archives/99e0fad2-1191-44b7-9f78-4f10a8594c99/data
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity must be linked to an account and the archive must belong to this account
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.

## Response (Success)

```bash
HTTP 200 OK
Content-Type: application/json; charset=UTF-8

"(a JSON string that should be a stringified JSON object)"
```

## Response (Archive Deleted)

```bash
HTTP 410 Gone
```

# Deleting a Backup Archive (After Recovery or on User Request)

```bash
DELETE https://api.misakey.com.local/backup-archives/3beb62c1-b11a-4ad6-9830-380ded30afa7
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 2): the identity must be linked to an account and the archive must belong to this account
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks. Delivered at the end of the auth flow.


```json
{
    "reason": "(string)",
}
```

`reason` must be either `"recovery"` (if you are deleting the archive because it was successfully recovered)
or `"deletion"` (if the user requested deletion of the archive).

## Response

```bash
HTTP 204 No Content
```
