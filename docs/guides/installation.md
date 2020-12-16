---
title: Installation
---

Misakey tech is distributed as a set of Docker images. This makes it easy to deploy it in various environments.

## Prerequisites

- A reachable SQL Database (PostgreSQL is recommended)
- A Key-Value store for caching (Redis is recommended)
- A S3 Compliant object storage
- A transactional email provider, or SMTP server to send emails

### Optional

- A Datadog account and agents to monitor your infrastructure

## Deploy with `docker-compose`

Running on docker-compose is quite easy. The recommended architecture is the same as the one described in [guide for running locally](/getting-started/running-locally.md).

You can take inspiration from the `docker-compose.yml` file in the guide and adjust it to your needs:
- Using your DB, Cache, S3 and email providers instead of local ones
- Managing your configuration files
- Using a `docker-compose` file version 3 to have something more production oriented (and being able to manage redundancy for instance).

## Deploy on Kubernetes

:::info

This section is a work in progress. We know we should add information about deployment and provide helm repos.
:::

If you want to deploy on Kubernetes, we provide some helm charts alongside our Docker images to make it easy to deploy on your cluster.

You can check on every code repository to get the dedicated helm chart to be able to deploy all services. The complementary information (config, secrets, ...) are available on each repository(frontend, backend/api, backend/gateway).

