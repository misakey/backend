import os

from . import http, URL_PREFIX
from .utils.base64 import urlsafe_b64encode

class Session(http.Session):
    def __init__(self):
        super().__init__()
        self.identity_id = ""
        self.email = ""

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

    def join_box(self, box_id):
        return self.post(
            f'{URL_PREFIX}/boxes/{box_id}/events',
            json={
                'type': 'member.join',
                'extra': {
                    # will be mandatory soon
                    # TODO take value as input argument
                    'other_share_hash': urlsafe_b64encode(os.urandom(16))
                }
            },
        )
