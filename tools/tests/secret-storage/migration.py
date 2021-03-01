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

def test_auth_datatags(user_session):
    print('- user creates an organization')
    r = user_session.post(
        f'{URL_PREFIX}/organizations',
        json={
            'name': 'Awesome Org'
        },
        expected_status_code=http.STATUS_CREATED
    )
    check_response(r,
        [
            lambda r: assert_fn(r.json()['name'] == 'Awesome Org'),
            lambda r: assert_fn(r.json()['creator_id'] == user_session.identity_id),
            lambda r: assert_fn(r.json()['current_identity_role'] == 'admin'),
        ]
    )
    created_org = r.json()

    print('- user sees self org and the created org')
    r = user_session.get(
        f'{URL_PREFIX}/identities/{user_session.identity_id}/organizations',
        expected_status_code=http.STATUS_OK
    )
    orgs = r.json() 
    check_response(r,
        [
            lambda r: assert_fn(len(orgs) == 2),
            lambda r: assert_fn(orgs[0]['id'] == SELF_CLIENT_ID),
            lambda r: assert_fn(orgs[1]['id'] == created_org['id']),
            lambda r: assert_fn(orgs[1]['name'] == created_org['name']),
            lambda r: assert_fn(orgs[1]['current_identity_role'] == 'admin'),
            lambda r: assert_fn(orgs[1]['creator_id'] == user_session.identity_id),
        ]
    )

    org_id = created_org['id']
    print(f'- user generates secret for the org {org_id}')
    r = user_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )

    print('- hydra client exists now for the organization with correct setup')
    r = user_session.get(
        f'{HYDRA_ADMIN_URL_PREFIX}/clients/{org_id}',
        expected_status_code=http.STATUS_OK
    )
    check_response(r,
        [
            lambda r: assert_fn(r.json()['client_id'] == org_id),
            lambda r: assert_fn(r.json()['grant_types'][0] == 'client_credentials'),
            lambda r: assert_fn(r.json()['token_endpoint_auth_method'] == 'client_secret_post'),
            lambda r: assert_fn(r.json()['audience'][1] == SELF_CLIENT_ID),
            lambda r: assert_fn(r.json()['scope'] == 'openid'),
            lambda r: assert_fn(r.json()['response_types'][0] == 'token'),
            lambda r: assert_fn(r.json()['subject_type'] == 'pairwise'),
        ]
    )

    print('- secret generation is idempotent')
    r = user_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )
    org_secret = r.json()['secret']

    print(f'- auth flow can be perform for the organization {org_id}')
    return get_org_session(org_id, org_secret)