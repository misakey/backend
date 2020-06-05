## Introduction

Nginx configuration to distribute requests across services.

## Expose a feature on the gateway

### Gateway Introduction

By default, the service is only reachable from `misakey_vpn` network, if you have configured it this way. In order to expose the service to the external world, we use a reverse proxy.

The reverse proxy we use today is [nginx](https://www.nginx.com/resources/wiki/).

This reverse proxy is responsible for:
- handling different domains of our set of products.
- serving frontend static files.
- serving some backend APIs:
  - it forwards requests to the corresponding service.

This being said, it is logical that a route exposed to the external world should always be exposed through the gateway.

### Where to configure my feature

If the feature to expose is frontend oriented, you should serve static files to corresponding domain (you might create one).
If the feature to expose is backend oriented, you might add it to `auth` business domain (auth.misakey.com) or our `generic api` business domain (api.misakey.com).

The `auth` business domain serves both frontend and backend routes, this is why the gateway consider backend only routes starting by an underscore `_` (we rewrite the location for every route).

The gateway should be rebuild when modified to have changes applied (`docker-compose up --build gateway`).

## Deploy the Gateway on Kubernetes

The gateway image is built and deployed each time the repo is pushed. It generates 3 images:

- On `tags`
  - `registry.misakey.dev/misakey/backend/gateway:<tag>`
- On `master`
  - `registry.misakey.dev/misakey/backend/gateway:preprod`
  - `registry.misakey.dev/misakey/backend/gateway:latest`


- Make sure the image you want to deploy exists
- In the `gateway` root directory, run:
  - If this is the first time
```
helm install --name nginx helm/gateway --set env=<preprod|production> --set image.tag=<preprod|tag> --set dns="<misakey.com|preprod.misakey.dev>"
```
  - If this is an upgrade
```
helm upgrade nginx helm/gateway --set env=<preprod|production> --set image.tag=<preprod|tag> --set dns="<misakey.com|preprod.misakey.dev>"
```
- Check that the deployment went well by running `kubectl get pods`.
