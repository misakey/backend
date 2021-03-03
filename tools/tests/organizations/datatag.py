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
    user_session = get_authenticated_session(acr_values=2)
    org_session = get_org_session(user_session)
    org_id = org_session.org_id

    print(f'- the access token can be used as bearer to list org datatags')
    r = org_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [lambda r: assert_fn(len(r.json())==0)])  


    print(f'- machine-org can create datatags')
    r = org_session.post(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        json={
            'name': 'contract'
        },
        expected_status_code=http.STATUS_CREATED
    )
    check_response(r, [
        lambda r: assert_fn(r.json()['id'] != ""),
        lambda r: assert_fn(r.json()['name'] == "contract"),
        lambda r: assert_fn(r.json()['organization_id'] == org_id),
    ])
    datatag_id = r.json()['id']
    r = org_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json())==1),
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
    ])

    print(f'- machine-org can edit an existing datatag')
    r = org_session.patch(
        f'{URL_PREFIX}/organizations/{org_id}/datatags/{datatag_id}',
        json={
            'name': 'pact'
        },
        expected_status_code=http.STATUS_NO_CONTENT
    )
    r = org_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json()) == 1),
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
        lambda r: assert_fn(r.json()[0]['name'] == "pact"),
    ])

    print(f'- admin can list org datatags')
    r = user_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [lambda r: assert_fn(len(r.json())==1)])  # 1 because machine-org has already created one

    print(f'- admin can create datatags')
    r = user_session.post(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        json={
            'name': 'salary'
        },
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(r.json()['id'] != ""),
        lambda r: assert_fn(r.json()['name'] == "salary"),
        lambda r: assert_fn(r.json()['organization_id'] == org_id),
    ])
    datatag_id = r.json()['id']
    r = user_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json())==2), # 2 because machine-org has already created one
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
    ])

    print(f'- admin can edit an existing datatag')
    r = user_session.patch(
        f'{URL_PREFIX}/organizations/{org_id}/datatags/{datatag_id}',
        json={
            'name': 'donation'
        },
        expected_status_code=http.STATUS_NO_CONTENT
    )
    r = user_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
        lambda r: assert_fn(r.json()[0]['name'] == "donation"),
    ])