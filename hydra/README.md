## Deploy Hydra on Kubernetes

We use the [official helm chart](https://k8s.ory.sh/helm/hydra).

1. `helm repo add ory https://k8s.ory.sh/helm/charts`.
2. `helm repo update`.
3. From the root of `hydra`, run `helm install --name=hydra -f hydra.<env>.yaml ory/hydra --set hydra.config.dsn=<postgres dsn> --set maester.enabled=false --set hydra.autoMigrate=true --set hydra.dangerousForceHttp=true --set hydra.config.oidc.subject_identifiers.pairwise.salt=<salt>`
4. To upgrade, use `helm upgrade hydra -f hydra.preprod.yaml ory/hydra --set hydra.config.dsn=<postgres_dsn> --set maester.enabled=false --set hydra.config.oidc.subject_identifiers.pairwise.salt=<salt> --set hydra.dangerousForceHttp=true --set hydra.config.secrets.system=<secret> --set image.tag=v1.5.2 --set hydra.autoMigrate=true --set replicaCount=2` for a potential image ugprade.
5. To make hydra accessible via our Traefik controller, we use a specific resource `IngressRoute` which is not in the Hydra helm chart. You need to install it from one of the files `ingress-route-<env>.yaml` with `kubectl apply -f`

Notes:

The secret is mandatory to re-enter otherwise client won't be able to authenticate to hydra (the is the first error you should encounter but not the only one).

The pairwise salt is mandatory to re-enter, otherwise hydra won't start by failing with an error:

`{"level":"fatal","msg":"The pairwise subject identifier algorithm was set but length of oidc.subject_identifier.salt is too small (0 \u003c 8), please set oidc.subject_identifiers.pairwise.salt to a random string with 8 characters or more.","time":"2020-01-08T13:06:17Z"}`
