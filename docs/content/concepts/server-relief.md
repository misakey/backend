---
title: Argon2 Server Relief (Client-side Password Hashing)
---

[Argon2][] is the password hashing function we use
to hash passwords before storing them in database.

Argon2 uses a lot of memory (by design)
and this results in our servers running out of memory
if they receive many login requests at the same time.

The solution to this is
to delegate the burden of hashing the password to the client
(the user's browser).
This is called *server relief*.

Server relief has another benefit:
the server never sees the password itself.
This is especially interesting for products like Misakey
where the user's password is also used to derive end-to-end encryption keys:
letting the server see the password unhashed
would almost be like letting the server see the keys,
so it would not really be end-to-end encryption any more.

In practice, server relief implies that
instead of sending a simple JSON string for the password
(`"password": "passw0rd123"`),
the client is expected to send an object with the following structure:

    {{% include "include/passwordHash.json" 4 %}}

[Argon2]: https://github.com/P-H-C/phc-winner-argon2