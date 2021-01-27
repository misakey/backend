---
title: Authentication
---

## Goal

Misakey lets you add modern authentication to your application in a few minutes. You have two possibilites to do so:

- You can host your own instance of Misakey SSO [locally](getting-started/running-locally.md) and [on your infrastructure](guides/deploy-on-prod.md)).
- You can use [Misakey's own SSO instance](https://app.misakey.com).

You can see an example of an app (an empty demo React application) using Misakey auth at https://demo.sso.misakey.com/.

## Integration
### Step 1: Create your SSO Client

If you are using your own instance of Misakey SSO, you must create your own sso client following [this guide](guides/create-auth-clients.md).

If you want to use [Misakey's instance](https://app.misakey.com), please contact us ([love@misakey.com](mailto:love@misakey.com)).
We will quickly create credentials on our instance and send them to you (the automated registration interface is still being developped).

### Step 2: Install the SDK

For now our SDK is only compatible with [React](https://reactjs.org/) applications. If you need another framework, you can tell us at ([love@misakey.com](mailto:love@misakey.com)).

```bash
npm install @misakey/sdk
```

### Step 3: Integrate the SDK in Your App

#### Import and Initialize the Configuration

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

#### Integrate the Hook in Your App

```js
const { isAuthCallback, isAuthenticated, userProfile } = useMisakeyAuth(authConfig);
if (isAuthCallback) { return null; }

```

#### Use the Auth Context

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

This SDK is still in beta version. If you need a more robust version of the SDK, contact us ([love@misakey.com](mailto:love@misakey.com)) to tell us what it is you need the most.

:::

