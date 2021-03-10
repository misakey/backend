#!/usr/bin/env python3
import json
import os
import random
import sys
from base64 import b64decode, b64encode
from time import sleep

from misapy import (AUTH_URL_PREFIX, HYDRA_ADMIN_URL_PREFIX, SELF_CLIENT_ID,
                    URL_PREFIX, http)
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.check_response import assert_fn, check_response
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session, get_org_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    # II - ACR 2 tests
    user_session = get_authenticated_session(acr_values=2)
    org_session = get_org_session(user_session)
    org_id = org_session.org_id

    print('- user sees self org and the created org')
    r = user_session.get(
        f'{URL_PREFIX}/identities/{user_session.identity_id}/organizations',
        expected_status_code=http.STATUS_OK
    )
    orgs = r.json() 
    check_response(r, [
            lambda r: assert_fn(len(orgs) == 2),
            lambda r: assert_fn(orgs[0]['id'] == SELF_CLIENT_ID),
            lambda r: assert_fn(orgs[1]['id'] == org_id),
            lambda r: assert_fn(orgs[1]['name'] == org_session.org_name),
            lambda r: assert_fn(orgs[1]['current_identity_role'] == 'admin'),
            lambda r: assert_fn(orgs[1]['creator_id'] == user_session.identity_id),
    ])

    print('- hydra client exists now for the organization with correct setup')
    r = user_session.get(
        f'{HYDRA_ADMIN_URL_PREFIX}/clients/{org_id}',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
            lambda r: assert_fn(r.json()['client_id'] == org_id),
            lambda r: assert_fn(r.json()['grant_types'][0] == 'client_credentials'),
            lambda r: assert_fn(r.json()['token_endpoint_auth_method'] == 'client_secret_post'),
            lambda r: assert_fn(r.json()['audience'][1] == SELF_CLIENT_ID),
            lambda r: assert_fn(r.json()['scope'] == 'openid'),
            lambda r: assert_fn(r.json()['response_types'][0] == 'token'),
            lambda r: assert_fn(r.json()['subject_type'] == 'pairwise'),
    ])

    print('- an identity exists for the organization')
    r = user_session.get(
        f'{URL_PREFIX}/identities/{org_id}/profile',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
            lambda r: assert_fn(r.json()['id'] == org_id),
            lambda r: assert_fn(r.json()['display_name'] == org_session.org_name),
    ])

    print('- secret generation is idempotent')
    r = user_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )

    # II - ACR 1 tests
    acr1_session = get_authenticated_session(acr_values=1)
    print('- acr 1 cannot create an organization')
    r = acr1_session.post(
        f'{URL_PREFIX}/organizations',
        json={
            'name': 'Awesome Org'
        },
        expected_status_code=http.STATUS_FORBIDDEN
    )
    check_response(r, [
            lambda r: assert_fn(r.json()['details']['acr'] == 'forbidden'),
            lambda r: assert_fn(r.json()['details']['required_acr'] == '2'),
    ])


    print('- acr 1 cannot list acr 2 orgs')
    acr1_session.get(
        f'{URL_PREFIX}/identities/{user_session.identity_id}/organizations',
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
    check_response(r, [
            lambda r: assert_fn(len(orgs) == 1),
            lambda r: assert_fn(orgs[0]['name'] == display_name),
            lambda r: assert_fn(orgs[0]['id'] == SELF_CLIENT_ID),
            lambda r: assert_fn(orgs[0]['creator_id'] == acr1_session.identity_id),
            lambda r: assert_fn(orgs[0]['current_identity_role'] == None),
    ])

    print('- acr 1 cannot generate secret for an org they do not own')
    acr1_session.put(
        f'{URL_PREFIX}/organizations/{org_session.org_id}/secret',
        expected_status_code=http.STATUS_FORBIDDEN
    )