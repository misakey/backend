import os
from uuid import uuid4

from .utils.base64 import urlsafe_b64encode

def random_secret_storage_reset_data():
    root_key = {
        "key_hash": urlsafe_b64encode(os.urandom(16)),
        "encrypted_key": urlsafe_b64encode(os.urandom(16))
    }
    vault_key = {
        "key_hash": urlsafe_b64encode(os.urandom(16)),
        "encrypted_key": urlsafe_b64encode(os.urandom(16))
    }
    # TODO make sure backend does not allow to reset crypto with zero asym keys,
    # since there must be at least the secret parts of identity keys
    asym_keys = { 
        urlsafe_b64encode(os.urandom(16)): {
            'encrypted_secret_key': urlsafe_b64encode(os.urandom(32)),
        },
        urlsafe_b64encode(os.urandom(16)): {
            'encrypted_secret_key': urlsafe_b64encode(os.urandom(32)),
        },
    }

    return {
        'account_root_key': root_key,
        'vault_key': vault_key,
        'asym_keys': asym_keys,
        # TODO check that backend returns an error if we don't provide this
        # on either account creation or password reset
        # (see https://gitlab.misakey.dev/misakey/backend/-/issues/294)
        'pubkey': urlsafe_b64encode(os.urandom(16)),
        'non_identified_pubkey': urlsafe_b64encode(os.urandom(16)),
        'pubkey_aes_rsa': (
            'com.misakey.aes-rsa-enc:' + urlsafe_b64encode(os.urandom(16))
        ),
        'non_identified_pubkey_aes_rsa': (
            'com.misakey.aes-rsa-enc:' + urlsafe_b64encode(os.urandom(16))
        ),
    }

def random_secret_storage_full_data():
    '''outputs random secret storage reset data
    plus some secrets that may have been collected before account creation'''
    reset_data = random_secret_storage_reset_data()

    box_key_shares = { 
        str(uuid4()): {
            'encrypted_invitation_share': urlsafe_b64encode(os.urandom(32)),
            'invitation_share_hash': urlsafe_b64encode(os.urandom(16)),
        },
        str(uuid4()): {
            'encrypted_invitation_share': urlsafe_b64encode(os.urandom(32)),
            'invitation_share_hash': urlsafe_b64encode(os.urandom(16)),
        },
    }

    return {
        **reset_data,
        'box_key_shares': box_key_shares,
    }