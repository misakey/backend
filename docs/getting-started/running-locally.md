---
title: Running locally
---

In this tutorial, you will get the Misakey's stack run locally in 5 min. After that you will be able to start hacking around. 

For production install, please check the dedicated guides for that.

## Install the stack

### Prerequisites

- Docker
- Docker compose
- Git
- Make

Add those lines to your `/etc/hosts` file:

```
127.0.0.1   api.misakey.com.local
127.0.0.1   auth.misakey.com.local
127.0.0.1   app.misakey.com.local
```

### Test & run project

We compiled everything to make the app runable locally in a project called [test and run](https://github.com/misakey/test-and-run).

Clone the repository with `git clone git@github.com:misakey/test-and-run.git`.

Run the command `make init`. It will initialize the local project, pull all the images, ... This can take a moment.

Edit the .env file to put the version of the stack you want to use. Current recommendation is:

```
NOTIFICATION_JOB_TAG=v0.8.1
FRONTEND_TAG=v1.8.0
GATEWAY_TAG=v0.0.6
API_TAG=v0.8.1
```

Launch the local application: `make application`

## Start trying the stack

Open your web browser, and go to [https://app.misakey.com.local](https://app.misakey.com.local).

:::tip
You will have to accept self-signed certificates (3 times): we decided to use SSL even in local env for more close to reality environment.
:::

Then you can play with the demo app: using the auth, using the data channels to discuss, ...

## Hacking around

From there, you can start to develop around the Misakey stack: creating your own app using the SSO and the other bricks of the system. 

You can find more complete documentation on the APIs and the usage of the stack in dedicated guides.