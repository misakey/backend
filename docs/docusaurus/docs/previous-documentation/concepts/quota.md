---
title: Storage quota
---

1. Concept

To handle different user plans in Misakey, we need to limit what the user can do in every plan and we have two notions for this: storage used and max authorized storage.

2. Objects 


#### Storage quotum

Storage quota are stored in `storage_quotum` database 

```
{
    "id": "<int>",
    "identity_id": "<uuid>",
    "value": "<int64>(bytes)",
    "origin": "base/etc"
}
```

Users can have several `storage_quotum` linked to their identityId.

The amount of storage available for a user is computed by retrieving all the `storage_quotum` object linked to the identity and sum all the `value` properties.

#### Used space

The amount of space consumed by a user is defined by:
 - the total size of text and file events contained in their own boxes (user is the creator of the box)
 - the total size of files contained in their vault

##### Box used space

The used space consumed by a box is depicted by `box_used_space` object:
```
{
    "id": "<int>",
    "box_id": "<uuid>",
    "value": "<int64>(bytes)",
}
```

The value of this objet is updated for every boxes each time a `create_event` of type `msg.file`, `msg.text`, `meg.edit` and `msg.delete` is performed.

##### Vault used space

The used space consumed by the vault is depicted by `vault_used_space` object:

```
{
    "value": "<int64>(bytes)"
}
```

The total space used by a user is obtained by summing all `box_used_space` values of boxes belonging to user, and adding `vault_used_space` value.
