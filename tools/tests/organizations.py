#!/usr/bin/env python3
import json
import os
import random
import sys
from base64 import b64decode, b64encode
from time import sleep

from misapy import URL_PREFIX, SELF_CLIENT_ID, http
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.check_response import assert_fn, check_response
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    acr2_session = get_authenticated_session(acr_values=2)
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