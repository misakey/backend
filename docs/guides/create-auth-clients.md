---
title: Create auth client
---

Before you can interact with Misakey tech, you have to create an authentication client to authenticate your users or give access to another service to your user data.

Misakey uses [ORY Hydra](https://www.ory.sh/docs/hydra) as an OpenID server. To create a client, you have to send a `POST` request to `http://your-hydra-admin-url:4445/clients` with the following body:

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
