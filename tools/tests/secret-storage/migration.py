#!/usr/bin/env python3

import os
from uuid import uuid4

from misapy.pretty_error import prettyErrorContext
from misapy.get_access_token import get_authenticated_session
from misapy import http, URL_PREFIX
from misapy.secret_storage import random_secret_storage_full_data
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
    payload = random_secret_storage_full_data()
    s1.post(
        f'{URL_PREFIX}/crypto/migration/v2',
        json=payload,
        expected_status_code=http.STATUS_NO_CONTENT
    )
    payload_without_id_keys = { **payload }
    del payload_without_id_keys['pubkey']
    del payload_without_id_keys['non_identified_pubkey']
    del payload_without_id_keys['pubkey_aes_rsa']
    del payload_without_id_keys['non_identified_pubkey_aes_rsa']
    r = s1.get(f'{URL_PREFIX}/crypto/secret-storage')
    check_response(
        r,
        [
            lambda r: assert_fn(struct_x_included_in_y(payload_without_id_keys, r.json())),
        ]
    )

    r = s1.get(f'{URL_PREFIX}/identities/{s1.identity_id}')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == payload['pubkey']),
            lambda r: assert_fn(r.json()['non_identified_pubkey'] == payload['non_identified_pubkey']),
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