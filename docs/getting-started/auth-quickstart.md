---
title: Authentication
---
## Goal

You can easily integrate authentication to your application using Misakey tech SSO.

You can host your own SSO instance (cf: the [guide to install the tech](guides/installation.md)), or use Misakey's SSO.

The result will be the integration of the auth system in your app. You can check our [SSO Demo](https://demo.sso.misakey.com/) to see the result on a fresh React app.

## Integration

### Step 1: create your SSO client

If you are using your own host of the Misakey tech SSO, please follow the [guide to create a client](guides/create-auth-clients.md).

If you want to use Misakey's SSO, please contact us ([love@misakey.com](mailto:love@misakey.com)) to have your credentials (the auto registration interface is still in development).

### Step 2: install the SDK

For now our SDK is compatible with React apps. If you need another framework, contact us ([love@misakey.com](mailto:love@misakey.com)) we will develop a SDK for your need.

```bash
npm install @misakey/sdk
```

### Step 3: integrate the SDK to your app

#### Import and initialize the configuration

```js
import { useMisakeyAuth } from "@misakey/sdk";

/** Copy this snippet in the config part of your app */
const authConfig = {
  clientId: 'e60ee766-d285-44a6-88c0-2d6e5c4633c1',
  redirectUri: 'http://localhost:3000/callback',
  buttonPlacement: 'top-right',
  userInfoRequirement: ['email'],
}

```

#### Integrate the hook in your app

```js
const { isAuthCallback, isAuthenticated, userProfile } = useMisakeyAuth(authConfig);
if (isAuthCallback) { return null; }

```

#### Use the auth context

```js
{isAuthenticated ? (
// Display your logged-in interface
`Hello ${userProfile.name}`
) : (
// Display your logged-out interface
'Hello stranger, please signin !'
)}
```

----

For more details, you can check out the [repository of the SDK](https://github.com/misakey/sso-js-sdk)

:::important

This SDK is still in beta. If you need a more robust version of the SDK, contact us ([love@misakey.com](mailto:love@misakey.com)) to give us feedback on the priority for you.

:::

