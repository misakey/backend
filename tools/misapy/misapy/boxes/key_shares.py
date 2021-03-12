import os

from ..utils.base64 import urlsafe_b64encode

def new_key_share_event():
    other_share_hash = urlsafe_b64encode(os.urandom(16))
    misakey_share = urlsafe_b64encode(os.urandom(16))
    encrypted_invitation_key_share = urlsafe_b64encode(os.urandom(32))

    return {
        'type': 'state.key_share',
        'extra': {
            'misakey_share': misakey_share,
            'other_share_hash': other_share_hash,
            'encrypted_invitation_key_share': encrypted_invitation_key_share,
        },
        # this event has no content
    }
