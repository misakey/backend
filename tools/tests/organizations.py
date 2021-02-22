#!/usr/bin/env python3
import json
import os
import random
import sys
from base64 import b64decode, b64encode
from time import sleep

from misapy import URL_PREFIX, AUTH_URL_PREFIX, HYDRA_ADMIN_URL_PREFIX, SELF_CLIENT_ID, http
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.check_response import assert_fn, check_response
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    # I - What is possible to do
    acr2_session = get_authenticated_session(acr_values=2)

    print('- acr 2 creates an organization')
    r = acr2_session.post(
        f'{URL_PREFIX}/organizations',
        json={
            'name': 'Awesome Org'
        },
        expected_status_code=http.STATUS_CREATED
    )
    check_response(r,
        [
            lambda r: assert_fn(r.json()['name'] == 'Awesome Org'),
            lambda r: assert_fn(r.json()['creator_id'] == acr2_session.identity_id),
            lambda r: assert_fn(r.json()['current_identity_role'] == 'admin'),
        ]
    )
    created_org = r.json()

    print('- acr 2 see self org and the created org')
    # then list orgs
    r = acr2_session.get(
        f'{URL_PREFIX}/identities/{acr2_session.identity_id}/organizations',
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
            lambda r: assert_fn(orgs[1]['creator_id'] == acr2_session.identity_id),
        ]
    )

    org_id = created_org['id']
    print('- acr 2 generate secret for the created org')
    r = acr2_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )

    print('- hydra client exists now for the organization with correct setup')
    r = acr2_session.get(
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
    r = acr2_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )
    secret = r.json()['secret']

    print('- auth flow can be perform for the organization')
    r = acr2_session.post(
        f'{AUTH_URL_PREFIX}/_/oauth2/token',
        data={
            'grant_type': 'client_credentials',
            'scope': '',
            'client_id': org_id,
            'client_secret': secret,
        },
    )

    # II - What is not possible to do
    
    acr1_session = get_authenticated_session(acr_values=1)
    print('- acr 1 cannot create an organization')
    r = acr1_session.post(
        f'{URL_PREFIX}/organizations',
        json={
            'name': 'Awesome Org'
        },
        expected_status_code=http.STATUS_FORBIDDEN
    )
    check_response(r,
        [
            lambda r: assert_fn(r.json()['details']['acr'] == 'forbidden'),
            lambda r: assert_fn(r.json()['details']['required_acr'] == '2'),
        ]
    )

    print('- acr 1 cannot list acr 2 orgs')
    r = acr1_session.get(
        f'{URL_PREFIX}/identities/{acr2_session.identity_id}',
        expected_status_code=http.STATUS_FORBIDDEN
    )

    print('- acr 1 only see self org')
    # first get the identity to compare display name and self org name
    r = acr1_session.get(
        f'{URL_PREFIX}/identities/{acr1_session.identity_id}',
        expected_status_code=http.STATUS_OK
    )
    display_name = r.json()['display_name']

    # then list orgs
    r = acr1_session.get(
        f'{URL_PREFIX}/identities/{acr1_session.identity_id}/organizations',
        expected_status_code=http.STATUS_OK
    )
    orgs = r.json()
    check_response(r,
        [
            lambda r: assert_fn(len(orgs) == 1),
            lambda r: assert_fn(orgs[0]['name'] == display_name),
            lambda r: assert_fn(orgs[0]['id'] == SELF_CLIENT_ID),
            lambda r: assert_fn(orgs[0]['creator_id'] == acr1_session.identity_id),
            lambda r: assert_fn(orgs[0]['current_identity_role'] == None),
        ]
    )

    print('- acr 1 cannot generate secret for the created org')
    r = acr1_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_FORBIDDEN
    )