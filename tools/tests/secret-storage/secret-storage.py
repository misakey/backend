#!/usr/bin/env python3

import os
import random
from uuid import uuid4

from misapy.pretty_error import prettyErrorContext
from misapy.get_access_token import get_authenticated_session, new_password_hash
from misapy import http, URL_PREFIX
from misapy.utils import struct_x_included_in_y
from misapy.utils.base64 import urlsafe_b64encode
from misapy.check_response import check_response, assert_fn
from misapy.password_hashing import hash_password
from misapy.get_access_token import get_authenticated_session

with prettyErrorContext():
    print('- account creation using secret storage')
    s1 = get_authenticated_session(acr_values=2, get_secret_storage=True)
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    root_key = r.json()['account_root_key']

    print('- update of encrypted root key when changing password')
    encrypted_account_root_key = urlsafe_b64encode(os.urandom(16))

    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/pwd-params')
    salt_base_64 = r.json()['salt_base_64']
    s1.put(
        f'{URL_PREFIX}/accounts/{s1.account_id}/password',
        json={
            'old_prehashed_password': hash_password('password', salt_base_64),
            'new_prehashed_password': new_password_hash('password'),
            'encrypted_account_root_key': encrypted_account_root_key,
        }
    )
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['account_root_key']['encrypted_key'] == encrypted_account_root_key),
            # account root key hash must not have changed
            lambda r: assert_fn(r.json()['account_root_key']['key_hash'] == root_key['key_hash']),
        ]
    )

    print('- reset of crypto when password is forgotten')
    # same account, different session
    # TODO change to reset_password=True (BROKEN for now)
    # s1 = get_authenticated_session(email=s1.email, reset_password=False)
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    # check_response(
    #     r,
    #     [
    #         lambda r: assert_fn(set(r.json().keys()) == {'account_root_key', 'vault_key', 'asym_keys', 'box_key_shares'}),
    #         lambda r: assert_fn(set(r.json()['account_root_key'].keys()) == {'encrypted_key', 'key_hash'}),
    #         lambda r: assert_fn(r.json()['account_root_key']['encrypted_key'] != encrypted_account_root_key),
    #         lambda r: assert_fn(r.json()['account_root_key']['key_hash'] != root_key['key_hash']),
    #         # lambda r: assert_fn(r.json()['vault_key'] != vault_key),
    #         # TODO test anything about "asym_keys" in the return value?
    #         lambda r: assert_fn(r.json()['box_key_shares'] == {}),
    #     ]
    # )
    root_key = r.json()['account_root_key']
    public_keys = set(r.json()['asym_keys'].keys())

    print('- "Forbidden" when adding asymmetric key with wrong root key hash')
    s1.post(
        f'{URL_PREFIX}/crypto/secret-storage/asym-keys',
        json={
            'public_key': urlsafe_b64encode(os.urandom(16)),
            'encrypted_secret_key': urlsafe_b64encode(os.urandom(16)),
            'account_root_key_hash': urlsafe_b64encode(os.urandom(16)),
        },
        expected_status_code=http.STATUS_FORBIDDEN,
    )

    print('- adding asymmetric key (success)')
    asym_key = {
        'public_key': urlsafe_b64encode(os.urandom(16)),
        'encrypted_secret_key': urlsafe_b64encode(os.urandom(16)),
        'account_root_key_hash': root_key['key_hash'],
    }
    s1.post(
        f'{URL_PREFIX}/crypto/secret-storage/asym-keys',
        json=asym_key,
        expected_status_code=http.STATUS_OK,
    )
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    new_pubkeys = set(r.json()['asym_keys'].keys()) - public_keys
    check_response(
        r,
        [
            lambda r: assert_fn(new_pubkeys == {asym_key['public_key']}),
            lambda r: assert_fn(r.json()['asym_keys'][asym_key['public_key']]['encrypted_secret_key'] == asym_key['encrypted_secret_key']),
        ]
    )

    print('- "Forbidden" when adding box key share key with wrong root key hash')
    s1.put(
        f'{URL_PREFIX}/crypto/secret-storage/box-key-shares/{str(uuid4())}',
        json={
            'invitation_share_hash': urlsafe_b64encode(os.urandom(16)),
            'encrypted_invitation_share': urlsafe_b64encode(os.urandom(32)),
            'account_root_key_hash': urlsafe_b64encode(os.urandom(16)),
        },
        expected_status_code=http.STATUS_FORBIDDEN,
    )

    print('- adding box key share (success)')
    box_key_share = {
        'invitation_share_hash': urlsafe_b64encode(os.urandom(16)),
        'encrypted_invitation_share': urlsafe_b64encode(os.urandom(32)),
        'account_root_key_hash': root_key['key_hash'],
    }
    box_id = str(uuid4())
    s1.put(
        f'{URL_PREFIX}/crypto/secret-storage/box-key-shares/{box_id}',
        json=box_key_share,
        expected_status_code=http.STATUS_OK,
    )
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    check_response(
        r,
        [
            lambda r: assert_fn(set(r.json()['box_key_shares'].keys()) == {box_id}), # there is just this box key share
            lambda r: assert_fn(struct_x_included_in_y(box_key_share, r.json()['box_key_shares'][box_id],
                                                       except_fields=['account_root_key_hash'])),
        ]
    )

    print('- cannot delete someone else\'s asym key')
    s2 = get_authenticated_session(acr_values=2)
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    pubkeys = set(r.json()['asym_keys'].keys())
    to_delete = random.choice(list(pubkeys))
    
    s2.delete(
        f'{URL_PREFIX}/crypto/secret-storage/asym-keys',
        json={
            'public_keys': [to_delete],
        },
        expected_status_code=http.STATUS_NOT_FOUND,
    )

    print('- deletion of asym keys')
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    pubkeys = set(r.json()['asym_keys'].keys())
    to_delete = random.choice(list(pubkeys))
    
    s1.delete(
        f'{URL_PREFIX}/crypto/secret-storage/asym-keys',
        json={
            'public_keys': [to_delete],
        },
        expected_status_code=http.STATUS_NO_CONTENT,
    )

    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    assert pubkeys - {to_delete} == set(r.json()['asym_keys'].keys())


    print('- cannot delete someone else\'s box key shares')
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    box_ids = set(r.json()['box_key_shares'].keys())
    to_delete = random.choice(list(box_ids))
    
    s2.delete(
        f'{URL_PREFIX}/crypto/secret-storage/box-key-shares',
        json={
            'box_ids': [to_delete],
        },
        expected_status_code=http.STATUS_NOT_FOUND,
    )

    print('- deletion of box key shares')
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    box_ids = set(r.json()['box_key_shares'].keys())
    to_delete = random.choice(list(box_ids))
    
    s1.delete(
        f'{URL_PREFIX}/crypto/secret-storage/box-key-shares',
        json={
            'box_ids': [to_delete],
        },
        expected_status_code=http.STATUS_NO_CONTENT,
    )

    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    assert box_ids - {to_delete} == set(r.json()['box_key_shares'].keys())
