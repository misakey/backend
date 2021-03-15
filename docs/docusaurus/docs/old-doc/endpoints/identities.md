---
title: Identities
---

# Introduction

## Concept

**Identity** is the core entity used for authorization and authentication at Misakey.

They are auto-generated: the system consider a new wild identity appears when an
unknown identifier is entered during a login flow.

They refer then to an identifier (email, phone number...) and they are optionally attached to [an account](/old-doc/endpoints/accounts.md).

End-users can claim an identity by creating an account on it or link an existing account.

They can be considered both as profiles and concrete identities. Users have many identities:
- as citizens: having both Korean and Russian nationalities...
- as internet fellows: having a Travian account and a Mastodon account...

## Resource Access

Resources access is based on the identity the end-user is logged in.

The end-user can select the identity they want to be logged in during the auth flow,
or in the interface at any moment.

There is no need to re-authenticate switching identities unless the security level
of an identity is higher than the current one the end-user is logged in.

# Base Identity

## Require an identity for a given identifier

Described in [the auth flow section](/old-doc/endpoints/auth_flow.md/#require-an-identity-for-a-given-identifier).

## Create an account on an identity

Described [in the accounts section](/old-doc/endpoints/accounts.md/#create-an-account-on-an-identity)

## Get an identity

This route allows the retrieval of the information related to an identity.

#### Request

```bash
GET https://api.misakey.com/identities/:id
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 200 CREATED
```

_JSON Body:_
```json
  {
    "id": "601611f3-5735-4039-8f64-8692aab68209",
    "account_id": "cbf5fd34-a4e0-4b69-8994-60eff9c1ada4",
    "identifier_value": "d9bdaa-test@misakey.com",
    "identifier_kind": "email",
    "display_name": "D9bdaa-Test",
    "notifications": "minimal",
    "avatar_url": null,
    "color": null,
    "level": 10,
    "mfa_method": "disabled",
    "pubkey": "6QvaldZMMtJdi1LUg4N0Ag",
    "non_identified_pubkey": "MUah4EnFPmyy6XA58WoG9A",
    "pubkey_aes_rsa": "com.misakey.aes-rsa-enc:dDLJjuwdcsTZIMJXsa6STg",
    "non_identified_pubkey_aes_rsa": "com.misakey.aes-rsa-enc:sCbt8_cgIxShuPHcKmRYrQ",
    "has_account": true,
    "has_totp_secret": false
  }

```

- `id` (uuid string): the unique identity id.
- `account_id` (uuid string) (nullable): the linked account unique id, always null if the end-user is connected with ACR 1.
- `has_account` (boolean): tell either the identity is linked or not to an account.
- `display_name` (uuid string): the name to display to represent the identity.
- `notifications` (uuid string): the frequency of notifications for this identity.
- `avatar_url` (uuid string) (nullable): the web-address of the avatar's file content.
- `identifier_value` (string): the value of the identifier.
- `identifier_kind` (string) (oneof: _email_): the kind of the identifier.
- `mfa_method` (string) (oneof: _disabled_, _totp_, _webauthn_): the mfa method used by the identity, default is `disabled`.

## Update an identity

Not all attributes can be updated (see below).

The request must be authenticated with a token corresponding to the updated identity.

#### Request

```bash
PATCH https://api.misakey.com/identities/:id
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

The fiels that can be patched are:
- `display_name` (string): the identity display name.
- `notifications` (string) (oneof: _minimal_, _moderate_, _frequent_): notification setting.
- `pubkey`, `non_identified_pubkey`, `pubkey_aes_rsa` and `non_identified_pubkey_aes_rsa`
- `mfa_method` (string) (oneof: _disabled_, _totp_, _webauthn_): configured mfa method of the user.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Upload an avatar

The request must be authenticated with a token corresponding to the identity on which the avatar is uploaded.

#### Request

```bash
PUT https://api.misakey.com/identities/:id/avatar
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Body Parameters (multipart/form\_data):_
- `avatar` (object): the avatar file.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Delete an avatar

The request must be authenticated with a token corresponding to the identity on which the avatar is deleted.

If no avatar is set on the identity, the request will return a `409 CONFLICT`.

#### Request

```bash
DELETE https://api.misakey.com/identities/:id/avatar
```
_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Getting All Identity Public Keys Associated to an Identifier

This must be used to build data for automatic invitations to boxes
(see [`access.add`-type events](/old-doc/concepts/box-events.md/#to-a-specific-identifier))

```bash
GET /identities/pubkey?identifier_value=michel@misakey.com
```

Success Response:

_Code:_
```bash
HTTP 200 OK
```

```json
[
  "urlSafeBase64PubKey",
  "anotherUrlSafeBase64PubKey"
]
```

# Identity Profiles

End-users can configure their identities to show more or less information publicly about them (email, phone number...).

By default, everything is hidden and only the display name is public.
Anyone (even not connected people) can access an identity profile page. Only the username is required to get it.

## Get an identity profile

#### Request

It is not mandatory to have any authorization nor session to call this endpoint.
If the call is authorized though, some information might be added to the profile.

As an example, admin of boxes may have more information than the default user profile. The profile owner might have consented to share their email in a box when the caller is an admin. In this case, the email is returned.

```bash
GET https://api.misakey.com/identities/:id/profile
```
_Cookies:_
- `accesstoken` (optional) (opaque token) (ACR >= 0): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 200 OK
```

```json
{
  "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
  "display_name": "Jean-Michel User",
  "avatar_url": null,
  "identifier_value": "jean-michel@misakey.com",
  "identifier_kind": "email",
  "non_identified_pubkey": "FYofPprIPU6qaHDtCNCETYtmmQQqdKvtJqYBF2pPXzc",
  "non_identified_pubkey_aes_rsa": "com.misakey.aes-rsa-enc:sCbt8_cgIxShuPHcKmRYrQ",
  "contactable": true
}
```

with attributes:
- `id`: (string, uuid) the unique id of the identity (can lead to the identity profile).
- `display_name`: (string) the display name of the sender.
- `avatar_url`: (string, nullable) the potential avatar url of the sender.
- `identifier_value`: (string, emptyable) the value of the identifier.
- `identifier_kind`, (string, emptyable, one of: email): the kind of the identifier.
- `non_identified_pubkey`, (string, emptyable) the public key of the identity.
- `contactable`, (bool) is the user directly contactable with Misakey?


## Configure the identity profile

The end-user can configure their identity profile they are connected on.
Using the request, they can enable/disable the visibility of some fields.

#### Request

Because of the request is a PATCH, each fields in body can be send alone or all together.

```bash
PATCH https://api.misakey.com/identities/:id/profile/config
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Body Parameters:_
```json
{
  "email": false,
}
```
with attributes:
- `email`: (mandatory because the only possibility today) (boolean): **true** to say it is shared publicly, **false** to make it private.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```

## Get the identity profile configuration

The end-user can see the field they have shared or not.

#### Request

```bash
GET https://api.misakey.com/identities/:id/profile/config
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_Body Parameters:_
```json
{
  "email": false,
}
```
with attributes:
- `email`: (boolean) **true** if shared publicly, **false** if private.

# Identity Notifications

Sometimes, the end-user must know some things happened that are linked to him:
- they have been kicked out a box: tell the end-user about it...
- they have reset their password: tell the implication of such a procedure for the crypto...
- they have just created their account: welcome the end-user...

This notifications are represented via a ressource on server side call identity notifications.

They can be acknowledged by their owner to let the system know they've seen it.

## Count unacknowldeged notifications for an identity

This request returns the number of identity notifications that have not been acknowledged yet.

#### Request

```bash
HEAD https://api.misakey.com/identities/:id/notifications
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```
_Headers:_
- `X-Total-Count` (integer): the total count of unacknowledged identity notifications.


## List notifications for an identity

This request returns the identity notification entities linked to an identity.
It handles pagination.

#### Request

```bash
GET https://api.misakey.com/identities/:id/notifications?offset=&limit=
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Query Parameters:_
- Pagination ([more info](/references/overview.mdx#pagination)). Default: infinite.


#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
[
  {
   "id": 5,
   "type": "user.create_identity", // the user has just created an ACR1 identity !
   "details": null,
   "created_at": "2038-11-04T23:59:03.000Z",
   "acknowledged_at": "2038-11-05T00:00:00.000Z",
  },
  {
   "id": 11,
   "type": "user.create_account", // the user has just created an account !
   "details": null,
   "created_at": "2038-11-05T00:00:02.000Z",
   "acknowledged_at": "2038-11-05T00:00:00.000Z",
  }, 
  {
   "id": 34,
   "type": "member.kick", // the user receiving this have been kicked out of a box
   "details": {
     "id": "e5d889de-6be1-4201-bb7e-0772fbbf41e2", // id of the concerned box
     "title": "Dossier client 33129" // title of the box
     "owner_org_id": "91ec8274-2b6d-40ff-afad-83e8ba5808e5", // owner org id of the box
   },
   "created_at": "2038-11-05T00:00:00.000Z",
   "acknowledged_at": "2038-11-05T00:00:05.000Z",
  }, 
  {
   "id": 87,
   "type": "user.reset_password", // the user has reset its password
   "details": null,
   "created_at": "2038-11-05T00:00:07.000Z",
   "acknowledged_at": null,
  }, 
  {
    "id": 113,
    "type": "box.auto_invite",
    "details": {
      "box_id": "50d03bf1-3515-43af-a3d9-d553d5b71dbc",
      "box_title": "Test Box",
      "owner_org_id": "91ec8274-2b6d-40ff-afad-83e8ba5808e5", // owner org id of the box
      "cryptoaction_id": "84cfa967-8cff-4ebc-ba0e-ce70ca700a79",
      "used": true, // if not used, attribute will not be present
    },
    "created_at": "2020-11-06T15:44:25.189269Z",
    "acknowledged_at": null
  }
]
```

with attributes for each object of the list:
- `id`: (integer) a unique integer corresponding to the identity notification.
- `type`: (string, one of: _member.kick_, _user.reset_password_, _user.create_account_, _user.create_identity_, _box.auto_invite_) the type of notification - details and displayed text should be set considering this value.
- `details`: (object) (nullable) a JSON object filled or `null` depending of the type of notification (see all JSON example to get info about it)
- `created_at`: (date) the moment the server created the notification.
- `acknowledged_at`: (date) (nullable) the moment the end-user has acknowledged the notification.

## Acknowledge all notifications for an identity

This request acknowledges all the current unacknowledged notifications for an identity.
A query parameter `ids` can be filled optionally to only acknowledge some specific notifications.

#### Request

```bash
PUT https://api.misakey.com/identities/:id/notifications/acknowledgement
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1): `mid` claim as the identity id.
- `tokentype` (optional): must be `bearer`.

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

_Path Parameters:_
- `id` (uuid string): the identity unique id.

_Query Parameters:_
- `ids` (string) (optional): coma-separated list of integer mentioning specific notifications to acknowledege, ex `34,35,65,1`. Ignored if not valid.

#### Response

_Code:_
```bash
HTTP 204 NO CONTENT
```
