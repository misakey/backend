```json
{
  "id": "fcfacf74-b15e-4583-bb71-55eb42cf2758",
  "display_name": "Jean-Michel User",
  "avatar_url": null,
  "identifier_id": "ae3ebd5e-b629-4e2f-9cd4-b2a69e2824b5",
  "identifier": {
      "value": "jean-michel@misakey.com",
      "kind": "email"
  },
  "non_identified_pubkey": "FYofPprIPU6qaHDtCNCETYtmmQQqdKvtJqYBF2pPXzc",
  "contactable": true,
}
```

with attributes:
- `id`: (string, uuid) the unique id of the identity (can lead to the identity profile).
- `display_name`: (string) the display name of the sender.
- `avatar_url`: (string, nullable) the potential avatar url of the sender.
- `identifier_id`: (string, uuid) the identifier id linked to the identity.
- `identifier.value`: (string, emptyable) the value of the identifier.
- `identifier.kind`, (string, emptyable, one of: email): the kind of the identifier.
- `non_identified_pubkey`, (string, emptyable) the public key of the identity.
- `contactable`, (bool) is the user directly contactable with Misakey?
