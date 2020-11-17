from . import http, URL_PREFIX

class Session(http.Session):
    def __init__(self, identity_id: str, email: str):
        super().__init__()
        self.identity_id = identity_id
        self.email = email

    def get_identity(self, id :str=None):
        return self.get(f'{URL_PREFIX}/identities/{id or self.identity_id}')

    def set_identity_pubkey(self, pubkey: str):
        return self.patch(
            f'{URL_PREFIX}/identities/{self.identity_id}',
            json={
                'pubkey': pubkey,
            },
        )

    def get_identity_pubkeys(self, identifier: str):
        return self.get(
            f'{URL_PREFIX}/identities/pubkey',
            params={
                'identifier_value': identifier,
            },
        )