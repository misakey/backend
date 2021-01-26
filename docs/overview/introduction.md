---
title: Introduction
slug: /
---

## Misakey in general

Misakey technology aims at building a standard for user management, personal data management and service interoperability for the modern Internet world.

The goal is to make it easy and quick to create web services and applications with a fluid user experience, a user-centric consent management and a high degree of interoperability.

Misakey tech is developed by the French company [Misakey](https://about.misakey.com). All the source code being developed is open source, and we are open to external contributions (see the [contribution guide](overview/contributing.md)).

## Misakey backend

Misakey backend is a service written in Golang that offers you user management and a secure data vault system through an HTTP API.

It is an Identity Provider with encrypted data exchanges built with it.
Both have been built together in order to have full control over authentication, authorizations strategies to make cryptographic concepts as smooth as possible to deal with, keeping security our main concern.

What does Misakey backend offer you when deployed:

- register/authenticate users using passwordless, password & MFA (TOTP/Webauthn) authentication methods.
- store any kind of user data using end-to-end encryption in dedicated vault.
- manage organizations and roles within the system.
- authorize organizations (humans and machines) and users to access encrypted vaults via user consent.

For doing so, backend uses:
- [OpenID Connect](https://openid.net/specs/openid-connect-core-1_0.html) (organizations are relying parties). backend uses the [Ory Hydra service](https://github.com/ory/hydra) that is necessary to deploy aside backend in your infra.
- [Cryptographic protocols](https://about.misakey.com/cryptography/white-paper.html).
- Databases (PostgreSQL, Redis)[^1].
- Emails (Amazon SES)[^1].
- File Storages (Amazon S3)[^1].

Incoming:
- More integrations[^1].
- HTTP SDKs to easily integrate the backend (React.js)[^1].
- Anything you think is good to add.

[^1]: we are waiting for you to request more integration/feature if required, please open a [Github issue](https://github.com/misakey/backend/issues/new) or contact us at [love@misakey.com](mailto:love@misakey.com).


## Next Section

In the next section, we will see the different components of the Misakey solution.