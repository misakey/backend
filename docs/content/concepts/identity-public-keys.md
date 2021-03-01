+++
categories = ["Concepts"]
date = "2020-09-11"
description = "Identity public keys"
tags = ["identity", "public", "keys"]
title = "Identity Public Keys"
+++

Identities have *two* public keys, a “non-identified” one and a normal (or “identified”) one. Actually they may have four of them since each type of public key can be set for each encryption algorithm (see [here](/concepts/encryption-algorithms))

The ”non-identified public key” is returned when requesting the public key of a particular identity, while the normal one is returned when requesting the public key associated to a particular identifier (recall, an identifier is typically an email).

We need these two different values because sometimes misakey users can find other users either by identity ID (profile page, for instance) or by identifier (when inviting people to a box, for instance) but we want to hide the association between an identity and its identifier. Such association can only be revealed to another Misakey user after the explicit consent of the owner of this identity. The reason is that Misakey users may not be willing to reveal the email address they use for their Misakey account, while at the same time email addresses are used for box access rules.