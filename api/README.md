# API service

API is the Misakey backend service

:warning: This section is a work in progress.

## Environment variables

- `ENV`: `production` or `development`. This will change some behaviours like the way to send emails
- `AWS_ACCESS_KEY`: Only on `production`. Needed to send emails.
- `AWS_SECRET_KEY`: Only on `production`. Needed to send emails.

## Migrations

The migrations steps are explained in each module.

## Deploy on Kubernetes

### Installation

**First** Create the config file, to get content (we will name it `config.prod.yaml` or `config.peprod.yaml` here), check on `/api/config/api.toml` for content. This should have this form:

```
config: >-
  [server]
    port = 5000
    ...
```

**Then** In the `api` root directory, run:

```
# For production
helm install --name api helm/api -f config.prod.yaml --set env=production,image.tag=vX.Y.Z,dns="misakey.com"

# For preprod
helm install --name api helm/api -f config.preprod.yaml --set env=preprod,image.tag=master,dns="preprod.misakey.dev",image.repository="registry.misakey.dev/misakey/backend/api"
```

### Upgrade

#### If there is a new configuration to deploy

**First** Create the config file, to get content, get current content online with `kubectl describe configmap api`, then past `api-config.toml` content in your local config file. This should have this form:

```
config: >-
  [server]
    port = 5000
    ...
```

**Then** In the `api` root directory, run:

```
# For production
helm upgrade api helm/api -f config.prod.yaml --set env=production,image.tag=vX.Y.Z,dns="misakey.com"

# For preprod
helm upgrade api helm/api -f config.preprod.yaml --set env=preprod,image.tag=master,dns="preprod.misakey.dev",image.repository="registry.misakey.dev/misakey/backend/api
```

#### If there is not a new configuration to deploy

```
# For production
helm upgrade api helm/api --reuse-values --set image.tag=vX.Y.Z"

# For preprod
helm upgrade api helm/api --reuse-values --set image.tag=master
```

