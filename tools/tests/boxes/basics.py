#!/usr/bin/env python3

import os
from base64 import b64encode

from misapy import http, URL_PREFIX, SELF_CLIENT_ID
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.boxes.key_shares import new_key_share_event
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)

    box1_id, box1_share_hash = create_box_and_post_some_events_to_it(session=s1)    
    create_box_and_post_some_events_to_it(session=s1)
    create_box_and_post_some_events_to_it(session=s1)

    print("- box key share retrieval")
    r = s1.get(f'{URL_PREFIX}/box-key-shares/{box1_share_hash}')
    check_response(r,[lambda r: assert_fn(r.json()["other_share_hash"] == box1_share_hash)])

    print("- get box public information")
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box1_id}/public?other_share_hash={box1_share_hash}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['owner_org_id'] == SELF_CLIENT_ID
    assert r.json()['creator']['display_name'] == s1.display_name
    assert r.json()['creator']['id'] == s1.identity_id
    assert r.json()['creator']['identifier_value'] == ''


    print("- get box")
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box1_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['owner_org_id'] == SELF_CLIENT_ID
    assert r.json()['creator']['display_name'] == s1.display_name
    assert r.json()['creator']['id'] == s1.identity_id
    assert r.json()['public_key'] != ''
    assert r.json()['access_mode'] == 'public'
    assert r.json()['last_event'] != None
    assert r.json()['settings'] != None
    assert r.json()['events_count'] != None

    print('- file upload')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/encrypted-files',
        files={
            'encrypted_file': os.urandom(64),
            'msg_encrypted_content': (None, b64encode(os.urandom(32)).decode()),
            'msg_public_key': (None, b64encode(os.urandom(32)).decode()),
        },
        expected_status_code=201,
    )

    print('- forbidden is returned while posting event on unexisting box id')
    r = s1.post(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=403
    )

    print('- forbidden is returned while getting a box with a non-existing id')
    r = s1.get(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922',
        expected_status_code=403,
    )

    print('- non-uuid box in path')
    r = s1.post(
        f'{URL_PREFIX}/boxes/YOU_KNOW_IM_BAD/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=400
    )

    print('- incorrect event content format')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.text',
        },
        expected_status_code=400
    )

    print('- pagination')
    r = s1.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'offset': 1,
            'limit': 2,
        }
    )
    boxes = r.json()
    assert len(boxes) == 2

    r = s1.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'limit': 10,
        }
    )
    boxes = r.json()
    # 3 boxes in total
    assert len(boxes) == 3

    r = s1.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'offset': 20,
            'limit': 2,
        }
    )
    boxes = r.json()
    assert boxes == []