#!/usr/bin/env python3
import json
import os
import random
import sys
from base64 import b64decode, b64encode
from time import sleep

from misapy import (AUTH_URL_PREFIX, HYDRA_ADMIN_URL_PREFIX, SELF_CLIENT_ID,
                    URL_PREFIX, http)
from misapy.box_helpers import create_org_box
from misapy.check_response import assert_fn, check_response
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session, get_org_session
from misapy.org_helpers import create_datatag
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    user_session = get_authenticated_session(acr_values=2)
    org_session = get_org_session(user_session)
    org_id = org_session.org_id
    datatag_id = create_datatag(org_session, org_id)

    print("- org creates a box")
    r = org_session.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={user_session.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]

    creation_box_body = {
        'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
        'title': 'Test Box',
        'owner_org_id': org_id,
        'datatag_id': datatag_id,
        'data_subject': user_session.email,
        'invitation_data': {
            pubkey: 'ShouldBeUnpaddedUrlSafeBase64'
        }
    }
    r = org_session.post(
        f'{URL_PREFIX}/organizations/{org_id}/boxes',
        json=creation_box_body,
        expected_status_code=201,
    )
    check_response(r, [
        lambda r: assert_fn(r.json()['title'] == 'Test Box'),
        lambda r: assert_fn(r.json()['owner_org_id'] == org_id),
        lambda r: assert_fn(r.json()['datatag_id'] == datatag_id),
        lambda r: assert_fn(r.json()['data_subject'] == user_session.email),
        lambda r: assert_fn(r.json()['public_key'] == 'ShouldBeUnpaddedUrlSafeBase64'),
        lambda r: assert_fn(r.json()['access_mode'] == 'limited'),
    ])
    box_id = r.json()['id']

    print("- org cannot use many endpoints")
    org_session.get(f'{URL_PREFIX}/boxes/{box_id}', expected_status_code=401)
    org_session.post(f'{URL_PREFIX}/boxes', json=creation_box_body, expected_status_code=401)

    print("- org retrieves that box")
    r = org_session.get(
        f'{URL_PREFIX}/organizations/{org_id}/boxes/{box_id}',
        expected_status_code=200,
    )
    check_response(r, [
        lambda r: assert_fn(r.json()['title'] == 'Test Box'),
        lambda r: assert_fn(r.json()['owner_org_id'] == org_id),
        lambda r: assert_fn(r.json()['datatag_id'] == datatag_id),
        lambda r: assert_fn(r.json()['data_subject'] == user_session.email),
        lambda r: assert_fn(r.json()['public_key'] == 'ShouldBeUnpaddedUrlSafeBase64'),
        lambda r: assert_fn(r.json()['access_mode'] == 'limited'),
    ])

    print("- org creates msg.text on box")
    r = org_session.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=201
    )
    print("- org creates msg.file on box")
    r = org_session.post(
        f'{URL_PREFIX}/boxes/{box_id}/encrypted-files',
        files={
            'encrypted_file': os.urandom(64),
            'msg_encrypted_content': (None, b64encode(os.urandom(32)).decode()),
            'msg_public_key': (None, b64encode(os.urandom(32)).decode()),
        },
        expected_status_code=201
    )
