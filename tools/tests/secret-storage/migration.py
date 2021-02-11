#!/usr/bin/env python3

import os
from uuid import uuid4

from misapy.pretty_error import prettyErrorContext
from misapy.get_access_token import get_authenticated_session
from misapy import http, URL_PREFIX
from misapy.utils import struct_x_included_in_y
from misapy.utils.base64 import urlsafe_b64encode
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session

with prettyErrorContext():
    # simulating not-yet-migrated account
    s1 = get_authenticated_session(acr_values=2, use_secret_backup=True)

    print('- indicate non-migrated accounts')
    s1.get(
        f'{URL_PREFIX}/crypto/secret-storage',
        expected_status_code=http.STATUS_CONFLICT
    )

    print('- account migration')
    root_key = {
        "key_hash": urlsafe_b64encode(os.urandom(16)),
        "encrypted_key": urlsafe_b64encode(os.urandom(16))
    }
    vault_key = {
        "key_hash": urlsafe_b64encode(os.urandom(16)),
        "encrypted_key": urlsafe_b64encode(os.urandom(16))
    }
    asym_keys = { 
        urlsafe_b64encode(os.urandom(16)): {
            'encrypted_secret_key': urlsafe_b64encode(os.urandom(32)),
        },
        urlsafe_b64encode(os.urandom(16)): {
            'encrypted_secret_key': urlsafe_b64encode(os.urandom(32)),
        },

    }
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
    identity_public_key = urlsafe_b64encode(os.urandom(16))
    identity_non_identified_public_key = urlsafe_b64encode(os.urandom(16))
    s1.post(
        f'{URL_PREFIX}/crypto/migration/v2',
        json={
            "account_root_key": root_key,
            'vault_key': vault_key,
            'asym_keys': asym_keys,
            'box_key_shares': box_key_shares,
            # backend doesn't check that these public keys are present in the "asym_keys" part
            'identity_public_key': identity_public_key,
            'identity_non_identified_public_key': identity_non_identified_public_key,
        },
        expected_status_code=http.STATUS_NO_CONTENT
    )
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['account_root_key'] == root_key),
            lambda r: assert_fn(r.json()['vault_key'] == vault_key),
            lambda r: assert_fn(r.json()['asym_keys'] == asym_keys),
            lambda r: assert_fn(struct_x_included_in_y(box_key_shares, r.json()['box_key_shares'])),
        ]
    )

    r = s1.get(f'{URL_PREFIX}/identities/{s1.identity_id}')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == identity_public_key),
            lambda r: assert_fn(r.json()['non_identified_pubkey'] == identity_non_identified_public_key),
        ]
    )
