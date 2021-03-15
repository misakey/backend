with attributes:
- `id`: (string, uuid) the unique id of the organization.
- `name`: (string) the name of the organization.
- `current_identity_role`: (string) (nullable) (one of: _admin_) the role for the current identity for this organization. _null_ is no special role attributed.
- `creator_id`: (string, uuid) the id of the identity who has created the organization.
- `created_at`: (date) the date of creation of the org.