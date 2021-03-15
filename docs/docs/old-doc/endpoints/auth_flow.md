---
title: Auth Flow
---

# Introduction

Performing an auth flow is the only to obtain an access token and ID token.
This section provides all the routes the frontend SSO application might use.

All routes described in the sequence diagram are not specified in this document.
Most of them do no reuqire specific frontend logic. Only the user's agent redirect
is required.

Routes are ordered consedering the expected way they should be called.

An auth flow is linked to both authorization and authentication.
A description of these concepts can be found in the [Authorization & Authentication specification](/old-doc/concepts/authzn.md).
It is probably worth to read before implement following routes.


**Authentication information**:

The final ID Token contains information about the authentication performed by the user.

See [the list of **Authentication Method References** and corresponding **Authentication Context Classes**](/old-doc/concepts/authzn.md/#methods) for more info.

## Overall auth flow

As of today:
- `app.misakey.com`: the frontend client
- `auth.misakey.com/_`: the Ory Hydra service
- `api.misakey.com`: the backend service responsible for authentication


[![diagram](https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-YXV0aC5taXNha2V5LmNvbS9fOiBpbml0aWF0ZXMgb2F1dGgyIGF1dGhvcml6YXRpb24gY29kZVxuICAgIGF1dGgubWlzYWtleS5jb20vXy0-PithcHAubWlzYWtleS5jb206IHJlZGlyZWN0cyB0aGUgdXNlcidzIGFnZW50IHdpdGggbG9naW4gY2hhbGxlbmdlXG4gICAgTm90ZSByaWdodCBvZiBhcHAubWlzYWtleS5jb206IFN0YXJ0cyBMb2dpbiBGbG93XG4gICAgYXBwLm1pc2FrZXkuY29tLS0-Pi1hcGkubWlzYWtleS5jb206IFxuICAgIGFwaS5taXNha2V5LmNvbS0tPmF1dGgubWlzYWtleS5jb20vXzogZmV0Y2hlcyBsb2dpbiBpbmZvXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-YXBpLm1pc2FrZXkuY29tOiBjaGVja3MgdXNlciBsb2dpbiBzZXNzaW9uc1xuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwcC5taXNha2V5LmNvbTogcmVkaXJlY3QgdG8gbG9naW4gcGFnZVxuICAgIGFwcC5taXNha2V5LmNvbS0-PmFwaS5taXNha2V5LmNvbTogcmVxdWlyZSBhbiBpZGVudGl0eSBmb3IgYW4gaWRlbnRpZmllclxuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwaS5taXNha2V5LmNvbTogcG90ZW50aWFsbHkgY3JlYXRlIGEgbmV3IGlkZW50aXR5L2FjY291bnRcbiAgICBhcGkubWlzYWtleS5jb20tPj5hcHAubWlzYWtleS5jb206IHJldHVybnMgdGhlIGlkZW50aXR5IGluZm9ybWF0aW9uXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-YXBpLm1pc2FrZXkuY29tOiBhdXRoZW50aWNhdGVzIHVzZXIgd2l0aCBjcmVkZW50aWFsc1xuICAgIGFwaS5taXNha2V5LmNvbS0tPmF1dGgubWlzYWtleS5jb20vXzogdHJhbnNtaXRzIGxvZ2luIGluZm8gYW5kIHJlY2VpdmVzIHJlZGlyZWN0IHVybCB3aXRoIGxvZ2luIHZlcmlmaWVyXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIGVuZCB1c2VyIHRvIGF1dGggc2VydmVyIHdpdGggbG9naW4gdmVyaWZpZXJcbiAgICBhcHAubWlzYWtleS5jb20tPj4tYXV0aC5taXNha2V5LmNvbS9fOiBcbiAgICBOb3RlIHJpZ2h0IG9mIGFwcC5taXNha2V5LmNvbTogRW5kcyBMb2dpbiBGbG93XG4gICAgYXV0aC5taXNha2V5LmNvbS9fLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHRoZSB1c2VyJ3MgYWdlbnQgd2l0aCBjb25zZW50IGNoYWxsZW5nZVxuICAgIGFwcC5taXNha2V5LmNvbS0-Pi1hcGkubWlzYWtleS5jb206IFxuICAgIE5vdGUgcmlnaHQgb2YgYXBwLm1pc2FrZXkuY29tOiBTdGFydHMgQ29uc2VudCBGbG93XG4gICAgYXBpLm1pc2FrZXkuY29tLS0-YXV0aC5taXNha2V5LmNvbS9fOiBmZXRjaGVzIGNvbnNlbnQgaW5mb1xuICAgIGFwaS5taXNha2V5LmNvbS0tPmFwaS5taXNha2V5LmNvbTogY2hlY2sgdXNlciBjb25zZW50IHNlc3Npb25zXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-YXBwLm1pc2FrZXkuY29tOiByZWRpcmVjdCB0byBjb25zZW50IHBhZ2VcbiAgICBhcHAubWlzYWtleS5jb20tPj5hcGkubWlzYWtleS5jb206IGNvbnNlbnQgdG8gc29tZSBzY29wZXNcbiAgICBhcGkubWlzYWtleS5jb20tLT5hdXRoLm1pc2FrZXkuY29tL186IHRyYW5zbWl0cyBjb25zZW50IGluZm8gYW5kIHJlY2VpdmVzIHJlZGlyZWN0IHVybCB3aXRoIGNvbnNlbnQgdmVyaWZpZXJcbiAgICBhcGkubWlzYWtleS5jb20tPj4rYXBwLm1pc2FrZXkuY29tOiByZWRpcmVjdHMgZW5kIHVzZXIgdG8gYXV0aCBzZXJ2ZXIgd2l0aCBjb25zZW50IHZlcmlmaWVyXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-LWF1dGgubWlzYWtleS5jb20vXzogXG4gICAgTm90ZSByaWdodCBvZiBhcHAubWlzYWtleS5jb206IEVuZHMgQ29uc2VudCBGbG93XG4gICAgYXBpLm1pc2FrZXkuY29tLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHVzZXIncyBhZ2VudCB0byByZWRpcmVjdCB1cmwgd2l0aCBjb2RlXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-LWFwaS5taXNha2V5LmNvbTogXG4gICAgYXBpLm1pc2FrZXkuY29tLS0-YXV0aC5taXNha2V5LmNvbS9fOiBmZXRjaGVzIHRva2VucyBhcyBhbiBhdXRoZW50aWNhdGVkIGNsaWVudFxuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHVzZXIncyBhZ2VudCB0byBmaW5hbCB1cmwgd2l0aCBJRCBUb2tlbiAoYW5kIGFjY2VzcyB0b2tlbiBhcyBjb29raWUpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-YXV0aC5taXNha2V5LmNvbS9fOiBpbml0aWF0ZXMgb2F1dGgyIGF1dGhvcml6YXRpb24gY29kZVxuICAgIGF1dGgubWlzYWtleS5jb20vXy0-PithcHAubWlzYWtleS5jb206IHJlZGlyZWN0cyB0aGUgdXNlcidzIGFnZW50IHdpdGggbG9naW4gY2hhbGxlbmdlXG4gICAgTm90ZSByaWdodCBvZiBhcHAubWlzYWtleS5jb206IFN0YXJ0cyBMb2dpbiBGbG93XG4gICAgYXBwLm1pc2FrZXkuY29tLS0-Pi1hcGkubWlzYWtleS5jb206IFxuICAgIGFwaS5taXNha2V5LmNvbS0tPmF1dGgubWlzYWtleS5jb20vXzogZmV0Y2hlcyBsb2dpbiBpbmZvXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-YXBpLm1pc2FrZXkuY29tOiBjaGVja3MgdXNlciBsb2dpbiBzZXNzaW9uc1xuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwcC5taXNha2V5LmNvbTogcmVkaXJlY3QgdG8gbG9naW4gcGFnZVxuICAgIGFwcC5taXNha2V5LmNvbS0-PmFwaS5taXNha2V5LmNvbTogcmVxdWlyZSBhbiBpZGVudGl0eSBmb3IgYW4gaWRlbnRpZmllclxuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwaS5taXNha2V5LmNvbTogcG90ZW50aWFsbHkgY3JlYXRlIGEgbmV3IGlkZW50aXR5L2FjY291bnRcbiAgICBhcGkubWlzYWtleS5jb20tPj5hcHAubWlzYWtleS5jb206IHJldHVybnMgdGhlIGlkZW50aXR5IGluZm9ybWF0aW9uXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-YXBpLm1pc2FrZXkuY29tOiBhdXRoZW50aWNhdGVzIHVzZXIgd2l0aCBjcmVkZW50aWFsc1xuICAgIGFwaS5taXNha2V5LmNvbS0tPmF1dGgubWlzYWtleS5jb20vXzogdHJhbnNtaXRzIGxvZ2luIGluZm8gYW5kIHJlY2VpdmVzIHJlZGlyZWN0IHVybCB3aXRoIGxvZ2luIHZlcmlmaWVyXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIGVuZCB1c2VyIHRvIGF1dGggc2VydmVyIHdpdGggbG9naW4gdmVyaWZpZXJcbiAgICBhcHAubWlzYWtleS5jb20tPj4tYXV0aC5taXNha2V5LmNvbS9fOiBcbiAgICBOb3RlIHJpZ2h0IG9mIGFwcC5taXNha2V5LmNvbTogRW5kcyBMb2dpbiBGbG93XG4gICAgYXV0aC5taXNha2V5LmNvbS9fLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHRoZSB1c2VyJ3MgYWdlbnQgd2l0aCBjb25zZW50IGNoYWxsZW5nZVxuICAgIGFwcC5taXNha2V5LmNvbS0-Pi1hcGkubWlzYWtleS5jb206IFxuICAgIE5vdGUgcmlnaHQgb2YgYXBwLm1pc2FrZXkuY29tOiBTdGFydHMgQ29uc2VudCBGbG93XG4gICAgYXBpLm1pc2FrZXkuY29tLS0-YXV0aC5taXNha2V5LmNvbS9fOiBmZXRjaGVzIGNvbnNlbnQgaW5mb1xuICAgIGFwaS5taXNha2V5LmNvbS0tPmFwaS5taXNha2V5LmNvbTogY2hlY2sgdXNlciBjb25zZW50IHNlc3Npb25zXG4gICAgYXBpLm1pc2FrZXkuY29tLT4-YXBwLm1pc2FrZXkuY29tOiByZWRpcmVjdCB0byBjb25zZW50IHBhZ2VcbiAgICBhcHAubWlzYWtleS5jb20tPj5hcGkubWlzYWtleS5jb206IGNvbnNlbnQgdG8gc29tZSBzY29wZXNcbiAgICBhcGkubWlzYWtleS5jb20tLT5hdXRoLm1pc2FrZXkuY29tL186IHRyYW5zbWl0cyBjb25zZW50IGluZm8gYW5kIHJlY2VpdmVzIHJlZGlyZWN0IHVybCB3aXRoIGNvbnNlbnQgdmVyaWZpZXJcbiAgICBhcGkubWlzYWtleS5jb20tPj4rYXBwLm1pc2FrZXkuY29tOiByZWRpcmVjdHMgZW5kIHVzZXIgdG8gYXV0aCBzZXJ2ZXIgd2l0aCBjb25zZW50IHZlcmlmaWVyXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-LWF1dGgubWlzYWtleS5jb20vXzogXG4gICAgTm90ZSByaWdodCBvZiBhcHAubWlzYWtleS5jb206IEVuZHMgQ29uc2VudCBGbG93XG4gICAgYXBpLm1pc2FrZXkuY29tLT4-K2FwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHVzZXIncyBhZ2VudCB0byByZWRpcmVjdCB1cmwgd2l0aCBjb2RlXG4gICAgYXBwLm1pc2FrZXkuY29tLT4-LWFwaS5taXNha2V5LmNvbTogXG4gICAgYXBpLm1pc2FrZXkuY29tLS0-YXV0aC5taXNha2V5LmNvbS9fOiBmZXRjaGVzIHRva2VucyBhcyBhbiBhdXRoZW50aWNhdGVkIGNsaWVudFxuICAgIGFwaS5taXNha2V5LmNvbS0-PmFwcC5taXNha2V5LmNvbTogcmVkaXJlY3RzIHVzZXIncyBhZ2VudCB0byBmaW5hbCB1cmwgd2l0aCBJRCBUb2tlbiAoYW5kIGFjY2VzcyB0b2tlbiBhcyBjb29raWUpIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0)

## Initiate an authorization code flow

#### Request

```bash
  GET https://auth.misakey.com/_/oauth2/auth
```

_Query Parameters:_
- see [Open ID Connect RFC](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).

#### Response

_Code:_
```bash
HTTP 302 FOUND
```

_Headers:_
- `Location`: https://api.misakey.com/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e

_HTML Body:_
```html
    <a href="https://api.misakey.com/auth/login?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e">Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

# Login Flow

## Get Login Information

This route is used to retrieve information about the current login flow using a login challenge.

#### Request

```bash
GET https://api.misakey.com/auth/login/info
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.

#### Response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "client": {
    "id": "cc411b8f-28bf-4d4e-abd9-99226b41da27",
    "name": "Misakey App",
    "logo_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png",
    "tos_uri": "https://about.misakey.com/#/fr/legals/tos/",
    "policy_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png"
  },
  "scope": [
    "openid"
  ],
  "acr_values": null,
  "login_hint": ""
}
```

- `client` (object): information about the SSO client involved in the auth flow:
  - `id` (uuid string): the unique id.
  - `name` (string): the name.
  - `logo_uri` (string) (nullable): web-address of the logo file.
  - `policy_uri` (string) (nullable): web-address of client privacy policy.
  - `tos_uri` (string) (nullable): web-address of the client TOS.
- `scope` (string): list of scope sent during the auth flow init.
- `acr_values` (string) (nullable): list of acr values sent during the auth flow init.
- `login_hint` (string): the login_hint sent during the auth flow init.

## Require an identity for a given identifier

This request is idempotent.

This route is used to retrieve information the identity the end-user will log in.

The identity can be a new one created for the occasion or an existing one.
See _Response_ below for more information.

#### Request

_Headers:_
- The request doesn't require an authorization header.

```bash
PUT https://api.misakey.com/auth/identities
```

_JSON Body:_
```json
{
	"login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
  "identifier_value": "auth@test.com",
  "password_reset": false
}
```

- `login_challenge` (string): can be found in preivous redirect URL.
- `identifier_value` (string): the identifier value the end-user has entered.
- `password_reset` (bool): a boolean to initiate a new password reset flow.

#### Response

This route returns the identity the end-user will login as.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "identity": {
    "display_name": "MUCHMICHMACH@test.com",
    "avatar_url": null,
    "has_account": true,
  },
  "authn_step": {
    "identity_id": "4a98b5a1-1c08-46c9-8f26-18d54cbed30a",
    "method_name": "(possibilities described below)",
    "metadata": (format described in the next section)
  },
}
```

- `identity` (object): the identity linked to the received identifier value.
  - `display_name` (string): a customizable display name.
  - `avatar_url` (string) (nullable): the web address of the end-user avatar file.
  - `has_account` (bool): if the identity has an account or not.
- `authn_step` (object): the preferred authentication step:
  - `identity_id` (uuid string): the unique identity id the authentication step is attached to.
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_, _account_creation): the preferred authentication method.
  - `metadata` (string) (nullable): filled considering the preferred/expected method.

### Metadata field formats as output

Considering the preferred/expected authentication method, the metadata on output can contain additional information.

:::tip
Metadata **as input** formats are defined [here](/old-doc/endpoints/auth_flow.md/#metadata-field-formats-as-input) and differ a bit.
:::

| Expected method formats descriptions :link: |
| ------------------------------------------------------------------------------------------- |
| [emailed_code](/old-doc/endpoints/auth_flow.md/#emailed_code-method-format-as-output) |
| [prehashed_password](/old-doc/endpoints/auth_flow.md/#prehashed_password-method-format-as-output) |
| [account_creation](/old-doc/endpoints/auth_flow.md/#account_creation-method-format-as-output) |
| [webauthn](/old-doc/endpoints/auth_flow.md/#webauthn-method-format-as-output) |
| [totp](/old-doc/endpoints/auth_flow.md/#totp-method-format-as-output) |
| [reset password](/old-doc/endpoints/auth_flow.md/#reset_password-method-format-as-output) |

#### emailed_code method format as output

```json
{
  [...]
  "method_name": "emailed_code",
  "metadata": null,
  [...]
}
```

#### prehashed_password method format as output

On `prehashed_password`, the `metadata` field contains information about how the password is supposed to be prehashed.

:warning: Warning, the metadata has not the exact same shape as [the metadata used to perform
an authentication step](/old-doc/concepts/authzn.md#possible-formats-for-the-metadata-field-1) using the `prehashed_password` method, which also contains the hash of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {
    "salt_base_64": "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==",
    "memory": 1024,
    "iterations": 1,
    "parallelism": 1
  },
  [...]
}
```

#### account_creation method format as output

```json
{
  [...]
  "method_name": "account_creation",
  "metadata": null,
  [...]
}
```

This is the only way to create an account linked to an identity.

:information_source: This method is returned by the server in a certain configuration.

This method is retured when:
- the final ACR is expected to be 2.
- the identity linked to the given identifier has no linked account.
- the end-user has provided a valid login session corresponding to a previous ACR 1 authentication.

:information_source: The authorization server might return first an `emailed_code` authn step in order to prove the user owns the identifier.

:::info
The `account_creation` will then come as a second authn step as a [More Authentication Required Response](/old-doc/endpoints/auth_flow.md/#the-more-authentication-required-response).
:::

:information_source: This step is skipped if the end-user has provided a valid login session corresponding to a previous ACR 1 authentication.

#### webauthn method format as output


```json
{
  [...]
  "method_name": "webauthn",
  "metadata":
      {
          "publicKey": {
              "challenge":"<string>",
              "timeout":<int>,
              "rpId":"<string>",
              "allowCredentials":[{
                  "type":"<string>",
                  "id":"<string>"
              },{
                  "type":"<string>",
                  "id":"<string>"
              }]
          }
    },
    [...]
}
```

The metadata content is defined and explained in the webauthn documentation.

#### totp method format as output


```json
{
  [...]
  "method_name": "totp",
  "metadata": null,
    [...]
}
```

#### reset_password method format as output

This method must only be used at the end of the flow.

This will set a new password for the given identity.

```json
{

  [...]
  "method_name": "reset_password",
  "metadata": null,
    [...]
}
```

## Perform an authentication step in the login flow

The next step to authenticate the end-user is to let them enter some information
assuring they own the identity. This is called an **authentication step**.

Some login flow will require many steps later but as of today, we only have one step
even for our most secure flows.

The metadata field contained in the authentication step depends of the method name.

#### Request

```bash
POST https://api.misakey.com/auth/login/authn-step
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_
```json
{
  "login_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "authn_step": {
    "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
    "method_name": "(possibilities defined in the next section)",
    "metadata": "(formats defined in the next section)"
  }
}
```

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the performed authentication step information:
  - `identity_id` (uuid string): the identity id.
  - `method_name` (string) (one of: _emailed\_code_, _prehashed\_password_, _account\_creation_, _webauthn_, _totp_): the authentication method used.
  - `metadata` (json object): metadata containing the emailed code value, the prehashed password or the webauthn options.
The list of possible formats is defined in the next section.

### Metadata field formats as input

This section describes the possible metadata format, as a JSON object, which is a
field contained in the JSON body of the previous section.

The context of this specification is the performing of an authentication step only.

:::tip
Metadata **as output** formats are described [here](/old-doc/endpoints/auth_flow.md/#metadata-field-formats-as-output) and differ a bit.
:::

| Expected method formats descriptions :link: |
| ------------------------------------------------------------------------------------------- |
| [emailed_code](/old-doc/endpoints/auth_flow.md/#emailed_code-method-format-as-input) |
| [prehashed_password](/old-doc/endpoints/auth_flow.md/#prehashed_password-method-format-as-input) |
| [account_creation](/old-doc/endpoints/auth_flow.md/#account_creation-method-format-as-input) |
| [webauthn](/old-doc/endpoints/auth_flow.md/#webauthn-method-format-as-input) |
| [totp](/old-doc/endpoints/auth_flow.md/#totp-method-format-as-input) |
| [reset password](/old-doc/endpoints/auth_flow.md/#reset_password-method-format-as-input) |


##### emailed_code method format as input

_JSON Body:_
```json
{
  [...]
  "method_name": "emailed_code",
  "metadata": {
    "code": "320028"
  },
  [...]
}
```

##### prehashed_password method format as input

:warning: Warning, the metadata has not the exact same shape as [the metadata returned requiring
an identity](/old-doc/concepts/authzn.md/#possible-formats-for-the-metadata-field) with the `prehashed_password` value as preferred method, which contains only the hash parameters of the password.

```json
{
  [...]
  "method_name": "prehashed_password",
  "metadata": {
    "hash_base_64": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
    "params": {
      "salt_base_64": "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==",
      "memory": 1024,
      "iterations": 1,
      "parallelism": 1
    }  
  },
  [...]
}
```

##### account_creation method format as input


```json
{
  [...]
  "method_name": "account_creation",
  "metadata": {
    "prehashed_password": {
      "hash_base_64": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
      "params": {
        "salt_base_64": "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==",
        "memory": 1024,
        "iterations": 1,
        "parallelism": 1
      }  
    },
    "secret_storage": {
      "account_root_key": {
          "key_hash": "ofUqfBb6u6mnU61XFYFBs4g",
          "encrypted_key": "SVqIxNjfDNSBLME0bTxBVg"
      },
      "vault_key": {
          "key_hash": "GRNiluqdewU0Deiw-GxDgQ",
          "encrypted_key": "RhkCn-l0OhPcRDqW4hFdeg"
      },
      "asym_keys": {
          "JTGeadip5O4-Hu6CgrndHA": {
              "encrypted_secret_key": "1dC9viP0rWxi6hPg1uKKQN9UhVYBUxebG_IV1cGCRYA"
          },
          "CppeRlQFRKn7yfQJArLEug": {
              "encrypted_secret_key": "mqmy4yZL-voAe0WxQRsO1ZofHUkpiz8y2nlaMoyKcrg"
          }
      },
      "pubkey": "6QvaldZMMtJdi1LUg4N0Ag",
      "non_identified_pubkey": "MUah4EnFPmyy6XA58WoG9A",
      "pubkey_aes_rsa": "com.misakey.aes-rsa-enc:dDLJjuwdcsTZIMJXsa6STg",
      "non_identified_pubkey_aes_rsa": "com.misakey.aes-rsa-enc:sCbt8_cgIxShuPHcKmRYrQ"
    },
  },
  [...]
}
```

##### webauthn method format as input


```json
{
  [...]
  "method_name": "webauthn",
  "metadata": {
      "id":"<string>",
      "rawId":"<string>",
      "response":{
          "clientDataJSON":"<string>",
          "authenticatorData":"<string>",
          "signature":"<string>",
          "userHandle":"<string>"
      },
      "type":"<string>"
    },
  [...]
}
```

The metadata content is explained in the webauthn documentation

##### totp method format as input


```json
{
  [...]
  "method_name": "totp",
  "metadata": {
      "code": "<string>",
      "recovery_code": "<string>"
    },
  [...]
}
```

The `code` must be the otp given by the external app.

The `recovery_code` must be in the user recovery codes set.

One of `code` or `recovery_code` is required, but the other one must be blank.

##### reset_password method format as input

_JSON Body:_
```json
  {
    "method_name": "reset_password",
    "metadata": {
      "prehashed_password": {
        "hash_base_64": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
        "params": {
          "salt_base_64": "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==",
          "memory": 1024,
          "iterations": 1,
          "parallelism": 1
        }
      },
      "secret_storage": {
        "account_root_key": {
            "key_hash": "oUqfBb6u6mnU61XFYFBs4g",
            "encrypted_key": "SVqIxNjfDNSBLME0bTxBVg"
        },
        "vault_key": {
            "key_hash": "GRNiluqdewU0Deiw-GxDgQ",
            "encrypted_key": "RhkCn-l0OhPcRDqW4hFdeg"
        },
        "asym_keys": {
            "JTGeadip5O4-Hu6CgrndHA": {
                "encrypted_secret_key": "1dC9viP0rWxi6hPg1uKKQN9UhVYBUxebG_IV1cGCRYA"
            },
            "CppeRlQFRKn7yfQJArLEug": {
                "encrypted_secret_key": "mqmy4yZL-voAe0WxQRsO1ZofHUkpiz8y2nlaMoyKcrg"
            }
        },
        "pubkey": "6QvaldZMMtJdi1LUg4N0Ag",
        "non_identified_pubkey": "MUah4EnFPmyy6XA58WoG9A",
        "pubkey_aes_rsa": "com.misakey.aes-rsa-enc:dDLJjuwdcsTZIMJXsa6STg",
        "non_identified_pubkey_aes_rsa": "com.misakey.aes-rsa-enc:sCbt8_cgIxShuPHcKmRYrQ"
      },
    },
  [...]
  }
```

- `password_reset` (object): information concerning the new password and backup.
  - `prehashed_password` (object): prehashed password using argon2:
    - `params` (object): argon2 parameters:
      - `memory` (integer).
      - `parallelism` (integer).
      - `iterations` (integer).
      - `salt_base64` (base64 string).
    - `hash_base64` (base64 string): the prehashed password.
  - `backup_data` (string): the new user backup data.

:::info
The `prehashed_password` contains information following [Argon2 server relief concepts(/old-doc/concepts/server-relief.md).
:::
#### Response

On success, the route can return two possible json body:

#### the "redirect" response

What is returned is the next URL the user's agent should be redirected to.
This response is given when the authentication server consider the end-user has proven its identity sufficiently.

The access token is sent and stored in an http-only cookie.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "next": "redirect",
  "redirect_to": "https://auth.misakey.com/_/oauth2/auth",
  "csrf_token": "3eb193b251a24dcb8bb5aa5c3cca3487"
}
```

- `next` (oneof: _redirect_, _authn_step_): the next action the authentication server is waiting for.
- `redirect_to` (string): the URL the user's agent should be redirected to.

#### The "more authentication required" response

What is returned is the next authentication step the end-user should perform.
This response is given when the authentication server requires more authentication step to proove the end-user identity.

Technically, it happens when the current ACR, corresponding to previously validated authentication steps (a.k.a. AMRs), is below the expected ACR.

The expected ACR can be set by:
- the Relying Party expectations: `acr_values` paramenter on the init of the auth flow.
- the choosen identity configuration.

Also part of the response, an access token that must be used for the next request as an `accesstoken` cookie. This token allows us to authorize more advanced calls.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "next": "authn_step",
  "authn_step": {
    "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
    "method_name": "...",
    "metadata": "..."
  }
}
```

- `next` (oneof: _redirect_, _authn_step_): the next action the authentication server is waiting for.
- `authn_step` (object): the next expected authn step to end the login flow.

:::info
**method_name** and **metadata** outputs are described [here](/old-doc/concepts/authzn.md/#/#metadata-field-formats-as-output).
:::

_Cookies_:
- `authnaccesstoken`: (string) an access token allowing more advanced requests while being still in the login flow.
- `authntokentype`: (string) the token type.

#### Notable error responses

On error during an authentication step, some information might be displayed to the end-user.

**1. Received code is invalid:**

This error occurs when the code received in metadata does not match any stored code.

_Code:_
```bash
HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
  "code": "forbidden",
  "origin": "body",
  "details": {
    "code": "invalid",
  },
}
```

**2. Received code has expired:**

This error occurs when the code received in metadata is correct but the timebox
to use it is expired.

_Code:_
```bash
HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
  "code": "forbidden",
  "origin": "body",
  "details": {
    "code": "expired",
  },
}
```

**3. The Authorization headers do not correspond to the login_challenge:**

Situation when the error is returned:
1. The end-user has performed an authentication step in a login flow A.
2. The end-user refreshes their agent, it inits a new login flow B.
3. The client has still the access token bound to the login flow A in headers.

The access token is valid for the login flow A but cannot be used in the login flow B.

_Code:_
```bash
HTTP 403 FORBIDDEN
```

_JSON Body:_
```json
{
  "code": "forbidden",
  "origin": "headers",
  "desc": "...",
  "details": {
    "Authorization": "conflict",
    "login_challenge": "conflict"
  }
}
```

## Init a new authentication step

This endpoint allows to init an authentication step:
- in case the last one has expired
- if a new step must be initialized

#### Request

```bash
POST https://api.misakey.com/authn-steps
```

```json
{
  "login_challenge": "e45f579fd02d41adbf8cb45e0f6a44ff",
  "authn_step": {
    "identity_id": "fed6784f-913b-49cb-9174-a8b7dc6bc675",
    "method_name": "emailed_code"
  }
}
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_

- `login_challenge` (string): can be found in previous redirect URL.
- `authn_step` (object): the initiated authentication step information:
  - `identity_id` (uuid string): the identity ID for which the authentication step will be initialized.
  - `method_name` (string) (one of: _emailed_code_, _prehashed_password_): the method used by the authentication step.

#### Response

This route does not return any content.

_Code:_
```bash
HTTP 204 NO CONTENT
```

#### Notable error responses

On errors, some information should be displayed to the end-user.

**1. A similar authn step already exists:**

This error occurs when an authentication step already exists for this `identity_id` and `method_name`

_Code:_
```bash
HTTP 409 Conflict
```

```json
{
  "code": "conflict",
  "origin": "body",
  "desc": "initing an authn step: a code has already been generated",
  "details": {
      "identity_id": "conflict",
      "method_name": "conflict"
  }
}
```

**2. Impossible to perform a prehashed_password method with the identity:**

This error occurs when the identity has no linked account. The password being attached to the account, such an authentication method is impossible to be handled.

_Code:_
```bash
HTTP 409 Conflict
```

```json
{
  "code": "conflict",
  "origin": "body",
  "desc": "initing an authn step: identity has no linked account",
  "details": {
      "identity_id": "conflict",
      "account_id": "required"
  }
}
```

# Consent Flow

## Get Consent Information

This route is used to retrieve information about the current consent flow using a consent challenge.

#### Request

```bash
GET https://api.misakey.com/auth/consent/info
```

_Query Parameters:_
- `consent_challenge` (string): the consent challenge corresponding to the current auth flow.

#### Response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "subject": "f058a198-e5e7-4d96-8b71-e6aa1edd3eb5",
  "acr": "2",
  "scope": [
    "openid",
    "tos",
    "privacy_policy"
  ],
  "context": {
    "amr": "emailed_code"
  },
  "client": {
    "id": "cc411b8f-28bf-4d4e-abd9-99226b41da27",
    "name": "Misakey App",
    "logo_uri": "https://media.glassdoor.com/sqll/2449676/misakey-squarelogo-1549446114307.png"
  }
}
```

- `subject` (uuid string): unique id of the account getting the token
- `acr` (string): the acr level for the current flow
- `scope` (string): list of scope sent during the auth flow init.
- `context` (object): context of the current consent flow
- `client` (object): information about the SSO client involved in the auth flow:
  - `id` (uuid string): the unique id.
  - `name` (string): the name.
  - `url` (string) (nullable): web-address of the logo file.


## Accept the consent request in the consent flow

This lets the user choose the scopes they want to accept.

For the moment, those scopes are limited to `tos` and `privacy_policy`

#### Request

```bash
POST https://api.misakey.com/auth/consent
```

_Headers:_
- The request doesn't require an authorization header.

_JSON Body:_
```json
{
  "consent_challenge": "e2645a0592e94ee78d8fbeaf65a4b82b",
  "identity_id": "53515d02-642a-4043-a943-bb11c0bdc6a5",
  "consented_scopes": [
    "tos",
    "privacy_policy"
  ]
}
```

- `consent_challenge` (string): can be found in previous redirect URL.
- `identity_id` (uuid string): the identity id bound to the identifier of the flow.
- `consented_scopes` (list of string) (one of: _tos_, _privacy\_policy_): the accepted scopes.

#### Response

On success, the route returns the next URL to redirect the user's agent.

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "redirect_to": "https://auth.misakey.com/_/oauth2/auth"
}
```

- `redirect_to` (string): the URL the user's agent should be redirected to.

#### Notable error responses

**1 - A mandatory scope is missing from consent**

If some legal scopes have been requested at the init of the auth flow, they
must be consented in all cases on this request.

Here is the error to expect if the client didn't send these scopes:
```json
{
  "code": "forbidden",
  "origin": "unknown",
  "details": {
    "requested_legal_scope": "{space-limited list of legal scope requested}",
    "consented_legal_scope": "{space-limited list of legal scope consented}"
  }
}
```

# Others

## Reset the auth flow

This requests allow the complete restart of the auth flow. It triggers a redirection to the initial
auth request if found (using the `login_challenge` sent in parameter).
If no auth request is found, it redirects the end-user to a blank connection screen on the Misakey main app (without any information about the initial flow).

:warning: be aware this action invalidates the session for the whole
account in this case.

#### Request

```bash
GET https://api.misakey.com/auth/reset?login_challenge=4f112272f2fa4cbe939b04e74dd3e49e
```

_Query Parameters:_
- `login_challenge` (string) (optional): the login challenge corresponding to the current auth flow. On invalid or missing, the user's agent will be redirected to the home page.

#### Response

_Code:_
```bash
HTTP 302 FOUND
```

_Headers:_
- `Location`: https://auth.misakey.com/_/oauth2/auth

_HTML Body:_
```html
    <a href="https://auth.misakey.com/_/oauth2/auth/>Found</a>
```

The `Location` header contains the same URL than the HTML body. The user's agent should be redirected to this URL to continue the auth flow to the login flow.

## Logout

This request logouts a user from their authentication session.

An authentication session is valid for an identity but it potentially links other identities
through the account relationship, be aware this action invalidates the session for the whole
account in this case.

#### Request

```bash
POST https://api.misakey.com/auth/logout
```


- `accesstoken` (opaque token) (ACR >= 0): `mid` claim as the identity id sent in body.
- `tokentype`: must be `bearer`

_Headers:_
- `X-CSRF-Token`: a token to prevent from CSRF attacks.

#### Response

This route does not return any content.

_Code:_
```bash
HTTP 204 NO CONTENT
```


## Get Secret Storage

This endpoint allows to get the account secret storage
during the auth flow.

This endpoint needs a valid process token.

#### Request

```bash
GET https://api.misakey.com.local/auth/secret-storage
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.
- `identity_id` (string) (uuid4): the id of the identity corresponding to the current auth flow.

_Headers_:
- `Authorization`: should be `Bearer {opaque_token}` with opaque token being the `login_challenge` of the auth flow.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
  "account_id": "d8aa7d0f-81fe-4e66-99d5-fe2b31360ae0",
  "secrets": {
    "account_root_key": {
        "key_hash": "V2Wrv7LjW_Ct_LPTuDcHSg",
        "encrypted_key": "JgdIJAGfImLyUlkk2eG5tQ"
    },
    "vault_key": {
        "key_hash": "JoUZHeHJ6zaz6vRw_Padgg",
        "encrypted_key": "RDTEMIlE4JlMAofbaF_VpQ"
    },
    "asym_keys": {
        "AbKozlB19lF8mGp0NURW0A": {
            "encrypted_secret_key": "DG5x6eRi9W56_BJpst5dgZCmu0O-s2gmkY_CPFkutF8"
        },
        "YobcbIi45V68XFOyU_q4nQ": {
            "encrypted_secret_key": "Tl53fZahxhijbuWH3qWkqOSDodt5UhlyNRVSyAegIpY"
        }
    },
    "box_key_shares": {}
  }
}
```

- `account_id` (string) (uuid4): the id of the account related to the current auth flow.
- `secrets` (object): the content of the account's secret storage.


## Get Backup

This endpoint allows to get the account backup
during the auth flow.

*(Note that secret backup system is not used anymore,
but this endpoint is needed for the frontend to migrate account still using it
to the new secret storage system)*

This endpoint needs a valid process token.

#### Request

```bash
GET https://api.misakey.com.local/auth/backup
```

_Query Parameters:_
- `login_challenge` (string): the login challenge corresponding to the current auth flow.
- `identity_id` (string) (uuid4): the id of the identity corresponding to the current auth flow.

_Headers_:
- `Authorization`: should be `Bearer {opaque_token}` with opaque token being the `login_challenge` of the auth flow.

#### Response

_Code:_
```bash
HTTP 200 OK
```

_JSON Body:_
```json
{
    "data": "[STRINGIFIED JSON]",
    "version": 3,
    "account_id": "d8aa7d0f-81fe-4e66-99d5-fe2b31360ae0"
}
```

- `data` (string): the user backup data.
- `version` (integer): the current backup version.
- `account_id` (string) (uuid4): the id of the account owning the backup.

## Creating a Root Key Share

This endpoint allows to create a root key share in the auth flow.

#### Request

```bash
  POST https://api.misakey.com/auth/root-key-shares
```

_Headers_:
- `Authorization`: should be `Bearer {opaque_token}` with opaque token being the access token given during the auth flow.

_JSON Body:_
```json
{
    "share": "o0hYlc2RurzJTiXldnnOMw",
    "other_share_hash": "axoGoSxJDiVWru3Sm-vdYQ"
}
```

- `share` (string) (base64): one of the shares.
- `other_share_hash` (string) (unpadded url-safe base64): a hash of the other share.

#### Response

_Code:_
```bash
HTTP 201 CREATED
```

_JSON Body:_
```json
{
  "account_id": "b2dc8b7e-44e6-4510-b222-c914876fad1c",
  "share": "o0hYlc2RurzJTiXldnnOMw",
  "other_share_hash": "axoGoSxJDiVWru3Sm-vdYQ"
}
```

# OIDC endpoints

These endpoints are openid RFC-compliant endpoints.

## Get User Info

This endpoint basically allow to get some of the ID token info.

It must be authenticated.

#### Request

```bash
GET https://api.misakey.com/auth/userinfo
```

_Cookies:_
- `accesstoken` (opaque token) (ACR >= 1)
- `tokentype`: must be `bearer`

#### Response

_Code_:
```bash
HTTP 200 OK
```

_JSON Body_:
```json
{
  "acr": "2",
  "aid": "97db9036-4190-4374-b0d1-0775b55f4e94",
  "amr": [
    "browser_cookie"
  ],
  "email": "joni@misakey.com",
  "mid": "aba11ab7-a077-4520-b76c-dba9fac01693",
  "sco": "openid tos privacy_policy",
  "sid": "562e4b78-86d2-4674-b80a-79dcfffb5d38",
  "sub": "2828b8d0-439c-4326-b49b-2736bd6eacb7"
}
```

- `acr` (string): ACR corresponding to the currrent token,
- `amr` (string array): the way the user got the token
- `email` (string): the user email,
- `sco` (string): list of space separated scopes linked to the token,
- `sid` (string): the session identifier,
- `sub` (string): the user identifier,
- `mid` (string): the Misakey specific identifier,
- `aid` (string): the user account identifier,
