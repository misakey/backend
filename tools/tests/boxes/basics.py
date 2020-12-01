#!/usr/bin/env python3

import os
from base64 import b64encode

from misapy import http, URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.boxes.key_shares import new_key_share_event
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    box1_id, box1_share_hash = create_box_and_post_some_events_to_it(session=s1, close=False)    

    print("- box key share retrieval")
    r = s1.get(f'{URL_PREFIX}/box-key-shares/{box1_share_hash}')
    check_response(r,[lambda r: assert_fn(r.json()["other_share_hash"] == box1_share_hash)])

    print("- get box public information")
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box1_id}/public?other_share_hash={box1_share_hash}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['creator']['display_name'] == s1.display_name
    assert r.json()['creator']['id'] == s1.identity_id
    assert r.json()['creator']['identifier']['value'] == ''

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

    print(f'- box closing')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'state.lifecycle',
            'content': {
                'state': 'closed'
            }
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
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
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

    print('- to create event on closed box is impossible')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=409
    )

    print('- non-creator cannot list events on a closed box')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        expected_status_code=403,
    )

    # Another identity creates other boxes
    box2_id, _ = create_box_and_post_some_events_to_it(session=s2, close=False)
    create_box_and_post_some_events_to_it(session=s2)
    create_box_and_post_some_events_to_it(session=s2)    

    print("- identity 1 becomes member of box 2")
    join_box(s1, box2_id)

    print('- identity 1 (creator) can list all events on open box 2')
    r = s1.get(f'{URL_PREFIX}/boxes/{box2_id}/events')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 4)
        ]
    )

    print('- identity 1 (non-creator) posts to box 2 a legit event')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box2_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        }
    )

    print('- identity 1 (non-creator) posts to box 2 a creator restricted event')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box2_id}/events',
        json={
            'type': 'state.lifecycle',
            'content': {
                'state': 'closed'
            }
        },
        expected_status_code=403
    )
    
    print('- identity 2 close the box and the identity 1 is notified')
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box2_id}/events',
        json={
            'type': 'state.lifecycle',
            'content': {
                'state': 'closed'
            }
        },
        expected_status_code=201,
    )
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications?offset=0&limit=2',
        expected_status_code=200
    )
    assert len(r.json()) == 2 # user.account_creation is also there
    assert r.json()[0]['type'] == 'box.lifecycle'
    assert r.json()[0]['details']['lifecycle'] == 'closed'
    assert r.json()[0]['details']['id'] == box2_id

    print(f'- boxes listing should return {box2_id} and {box1_id}')
    r = s1.get(f'{URL_PREFIX}/boxes/joined', expected_status_code=200)
    boxes = r.json()
    assert len(boxes) == 2
    
    # identity one did not take part into box 3 so it should not be returned
    assert set(map(lambda box: box['id'], boxes)) == {box1_id, box2_id}

    print('- pagination')
    r = s2.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'offset': 1,
            'limit': 2,
        }
    )
    boxes = r.json()
    assert len(boxes) == 2

    r = s2.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'offset': 1,
            'limit': 10,
        }
    )
    boxes = r.json()
    # Identity 2 has 3 boxes in total
    assert len(boxes) == 2

    r = s2.get(
        f'{URL_PREFIX}/boxes/joined',
        params={
            'offset': 20,
            'limit': 2,
        }
    )

    boxes = r.json()
    assert boxes == []
