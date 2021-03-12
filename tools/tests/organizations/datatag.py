#!/usr/bin/env python3
import json
import os
import random
import sys
from base64 import b64decode, b64encode
from time import sleep

from misapy import (AUTH_URL_PREFIX, HYDRA_ADMIN_URL_PREFIX, SELF_CLIENT_ID,
                    URL_PREFIX, http)
from misapy.box_helpers import create_box_and_post_some_events_to_it, create_org_box
from misapy.org_helpers import create_datatag, create_org
from misapy.check_response import assert_fn, check_response
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session, get_org_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    admin_session = get_authenticated_session(acr_values=2)
    user_session = get_authenticated_session(acr_values=2)
    org_session = get_org_session(admin_session)
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
    r = admin_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [lambda r: assert_fn(len(r.json())==1)])  # 1 because machine-org has already created one


    print(f'- admin can create datatags')
    r = admin_session.post(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        json={
            'name': 'salary'
        },
        expected_status_code=http.STATUS_CREATED
    )
    check_response(r, [
        lambda r: assert_fn(r.json()['id'] != ""),
        lambda r: assert_fn(r.json()['name'] == "salary"),
        lambda r: assert_fn(r.json()['organization_id'] == org_id),
    ])
    datatag_id = r.json()['id']
    r = admin_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json())==2), # 2 because machine-org has already created one
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
    ])

    print(f'- admin can edit an existing datatag')
    r = admin_session.patch(
        f'{URL_PREFIX}/organizations/{org_id}/datatags/{datatag_id}',
        json={
            'name': 'donation'
        },
        expected_status_code=http.STATUS_NO_CONTENT
    )
    r = admin_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        expected_status_code=http.STATUS_OK
    )
    check_response(r, [
        lambda r: assert_fn(r.json()[0]['id'] == datatag_id),
        lambda r: assert_fn(r.json()[0]['name'] == "donation"),
    ])

    print(f'- user can list datatags corresponding to them')
    org_session1 = get_org_session(admin_session)
    org_id1 = org_session1.org_id
    did_org1 = create_datatag(admin_session, org_id1)
    r = admin_session.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={user_session.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]

    box = create_org_box(org_session1, data_subject=user_session.email, org_id=org_id1, datatag_id=did_org1, public_key=pubkey)
    box_id = box['id']

    user_session.join_box(box_id)

    r = user_session.get(
            f'{URL_PREFIX}/identities/{user_session.identity_id}/datatags?organization_id={org_id1}',
            expected_status_code=200,
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json()) == 1),
        lambda r: assert_fn(r.json()[0]['id'] == did_org1),
        lambda r: assert_fn(r.json()[0]['organization_id'] == org_id1),
    ])

    print(f'- user have no datatags for organization with no boxes')
    no_box_org_id = create_org(admin_session)
    r = user_session.get(
            f'{URL_PREFIX}/identities/{user_session.identity_id}/datatags?organization_id={no_box_org_id}',
            expected_status_code=200,
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json()) == 0),
    ])

    print(f'- user do not retrieve datatags from other orgs')
    org_session2 = get_org_session(admin_session)
    org_id2 = org_session2.org_id
    did_org2 = create_datatag(admin_session, org_id2)
    r = admin_session.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={user_session.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]

    datatag_id = create_datatag(admin_session, org_id)
    box_org1 = create_org_box(org_session1, data_subject=user_session.email, org_id=org_id1, datatag_id=did_org1, public_key=pubkey)
    box_org2 = create_org_box(org_session2, data_subject=user_session.email, org_id=org_id2, datatag_id=did_org2, public_key=pubkey)

    user_session.join_box(box_org1['id'])
    user_session.join_box(box_org2['id'])

    r = user_session.get(
            f'{URL_PREFIX}/identities/{user_session.identity_id}/datatags?organization_id={org_id1}',
            expected_status_code=200,
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json()) == 1),
        lambda r: assert_fn(r.json()[0]['id'] == did_org1),
        lambda r: assert_fn(r.json()[0]['organization_id'] == org_id1),
    ])

    r = user_session.get(
            f'{URL_PREFIX}/identities/{user_session.identity_id}/datatags?organization_id={org_id2}',
            expected_status_code=200,
    )
    check_response(r, [
        lambda r: assert_fn(len(r.json()) == 1),
        lambda r: assert_fn(r.json()[0]['id'] == did_org2),
        lambda r: assert_fn(r.json()[0]['organization_id'] == org_id2),
    ])

    print(f'- user can’t list datatags of another user')
    r = user_session.get(
            f'{URL_PREFIX}/identities/{admin_session.identity_id}/datatags?organization_id={org_id1}',
            expected_status_code=403,
    )
