---
title: Create auth client
---

The totally first step to start an interaction with Misakey tech is to create an auth client to authenticate your users or give access to another service to your user data.

As the OpenID server, we are using [ORY Hydra](https://www.ory.sh/docs/hydra). To create a client, you have to:

`POST` on `http://your-hydra-admin-url:4445/clients`

```json
{
        "client_id": "DEFINE A UUIDv6",
        "client_name": "NAME OF YOUR CLIENT",
        "redirect_uris": ["https://MISAKEY_URL.your-org.tld/login/callback"],
        "grant_types": ["authorization_code", "client_credentials"],
        "response_types": ["id_token", "token", "code"],
        "scope": "openid email",
        "subject_type": "pairwise",
        "token_endpoint_auth_method": "client_secret_post",
        "userinfo_signed_response_alg": "none",
}
```
