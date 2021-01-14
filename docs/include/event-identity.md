with attributes:
- `id`: (string, uuid) the unique id of the identity (can lead to the identity profile).
- `display_name`: (string) the display name of the sender.
- `avatar_url`: (string, nullable) the potential avatar url of the sender.
- `identifier_value`: (string, emptyable) the value of the identifier.
- `identifier_kind`, (string, emptyable, one of: email): the kind of the identifier.