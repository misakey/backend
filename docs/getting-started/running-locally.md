---
title: Running locally
---

In this tutorial, you will get to run the Misakey's stack locally in 5 min. After that you will be able to start hacking around. 

For production install, please check [the dedicated guide](guides/installation.md).

## Install the Stack

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

### Misakey's “Test & Run” Project

We grouped all the tools required to run the app locally in a project called [test and run](https://github.com/misakey/test-and-run).

Clone the repository with `git clone git@github.com:misakey/test-and-run.git`.

Run the command `make init`. It will initialize the local project, pull all the images etc… This can take some time.

Edit the `.env` file to set the version of the stack you want to use. Current recommendation is:

```
NOTIFICATION_JOB_TAG=v0.8.1
FRONTEND_TAG=v1.8.0
GATEWAY_TAG=v0.0.6
API_TAG=v0.8.1
```

Launch the whole stack locally: `make application`

## Using the Application

Open your web browser and go to [https://app.misakey.com.local](https://app.misakey.com.local).

:::tip
You will have to accept self-signed certificates (3 times): we decided to use SSL even in local development environement to be closer to a real-world environment.
:::

Then you can play with the demo application: using the auth, creating data channels and chatting through them…

## Hacking Around

From there, you can create your own app using the SSO and the other bricks of the system. 

You can find a more complete documentation of the APIs and the usage of the stack in dedicated guides.