---
title: Running locally
---

In this tutorial, you will get to run the Misakey's stack locally in 5 min. After that you will be able to start hacking around. 

For production install, please check [the dedicated guide](guides/deploy-on-prod.md).

## Install the Stack

### Prerequisites

- Python3 (with Pip)
- Docker
- Docker compose
- Git

### Misakey's “Test & Run” Project

We grouped all the tools required to run the app locally in a project called [test and run](https://github.com/misakey/test-and-run).

Clone the repository with `git clone git@github.com:misakey/test-and-run.git`.

Install the CLI:
- Go to the `misacli` directory
- Make sure you have `pip` for Python 3
- Run `pip install -e .`

:warning: The `misacli` CLI must be used in the root directory of the project.

Run the command `misacli init` and follow the instructions.

The others commands are described in the CLI help (`misacli --help`).

**Example:** To run the whole application, run `misacli run app`.

## Using the Application

Open your web browser and go to [https://app.misakey.com.local](https://app.misakey.com.local).

:::tip
You will have to accept self-signed certificates (3 times): we decided to use SSL even in local development environement to be closer to a real-world environment.
:::

Then you can play with the demo application: using the auth, creating data channels and chatting through them…

## Hacking Around

From there, you can create your own app using the SSO and the other bricks of the system. 

You can find a more complete [documentation of the APIs](https://backend.docs.misakey.dev/) and the usage of the stack in dedicated guides.
