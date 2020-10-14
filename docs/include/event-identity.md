with attributes:
- `display_name`: (string) the display name of the sender.
- `avatar_url`: (string, nullable) the potential avatar url of the sender.
- `identifier_id`: (string, uuid) the identifier id linked to the identity.
- `identifier.value`: (string, emptyable) the value of the identifier.
- `identifier.kind`, (string, emptyable, one of: email): the kind of the identifier.