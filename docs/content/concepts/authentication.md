---
title: Authentication
---

All endpoints require authentication
unless explicitely mentionned otherwise.

When an endpoint requires authentication,
the request is expected to have the following HTTP header:

    Authorization: Bearer ACCESS_TOKEN

Where `ACCESS_TOKEN` is the access token obtained
through the authentication process.