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
from misapy.encryption import aes_rsa

with prettyErrorContext():
    user_session = get_authenticated_session(acr_values=2)
    org_session = get_org_session(user_session)
    org_id = org_session.org_id
    datatag_id = create_datatag(org_session, org_id)

    print("- org creates a box")
    box_keys = aes_rsa.generate_key_pair()
    r = org_session.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={user_session.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]

    creation_box_body = {
        'public_key': box_keys.public_key,
        'title': 'Test AES-RSA Box',
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
        lambda r: assert_fn(r.json()['title'] == 'Test AES-RSA Box'),
        lambda r: assert_fn(r.json()['owner_org_id'] == org_id),
        lambda r: assert_fn(r.json()['datatag_id'] == datatag_id),
        lambda r: assert_fn(r.json()['data_subject'] == user_session.email),
        lambda r: assert_fn(r.json()['public_key'] == box_keys.public_key),
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
        lambda r: assert_fn(r.json()['title'] == 'Test AES-RSA Box'),
        lambda r: assert_fn(r.json()['owner_org_id'] == org_id),
        lambda r: assert_fn(r.json()['datatag_id'] == datatag_id),
        lambda r: assert_fn(r.json()['data_subject'] == user_session.email),
        lambda r: assert_fn(r.json()['public_key'] == box_keys.public_key),
        lambda r: assert_fn(r.json()['access_mode'] == 'limited'),
    ])

    print("- org creates msg.text on box")
    r = org_session.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': aes_rsa.encrypt_message(b'HI FROM MISAPY!', box_keys.public_key),
                'public_key': box_keys.public_key,
            }
        },
        expected_status_code=http.STATUS_CREATED,
    )

    print("- org creates msg.file on box")
    encrypted_file, encrypted_msg_content = aes_rsa.encrypt_file(b'I AM A FILE', 'its_a_file.txt', box_keys.public_key)
    r = org_session.post(
        f'{URL_PREFIX}/boxes/{box_id}/encrypted-files',
        files={
            'encrypted_file': encrypted_file,
            'msg_encrypted_content': (None, encrypted_msg_content),
            'msg_public_key': (None, box_keys.public_key),
        },
        expected_status_code=http.STATUS_CREATED,
    )
