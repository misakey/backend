#!/usr/bin/env python3

import json
import os
import sys
from time import sleep
from base64 import b64encode, b64decode

from misapy import http
from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext
from misapy.container_access import list_encrypted_files
from misapy.check_response import check_response, assert_fn

URL_PREFIX = 'http://127.0.0.1:5020'


def create_box_and_post_some_events_to_it(session, close=True):
    s = session

    r = s.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        }
    )
    creator = r.json()['creator']
    assert creator['identifier']['value'] == s.email

    box_id = r.json()['id']

    print(f'Testing create msg.text event on box {box_id}')
    r = s.post(
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

    if close:
        print(f'Testing close box {box_id}')
        r = s.post(
            f'{URL_PREFIX}/boxes/{box_id}/events',
            json={
                'type': 'state.lifecycle',
                'content': {
                    'state': 'closed'
                }
            },
            expected_status_code=201
        )

    print(f'Testing listing for created box {box_id}')
    r = s.get(f'{URL_PREFIX}/boxes/{box_id}/events')
    event_list = r.json()
    assert (len(event_list) == 3 if close else 2)
    for event in event_list:
        assert 'id' in event
        assert 'server_event_created_at' in event
        assert 'type' in event
        assert 'content' in event

        sender = event['sender']
        assert sender['identifier']['value'] == s.email

    return box_id


with testContext():
    # Init 2 user sessions for creator rules testing
    s1 = get_authenticated_session()
    s2 = get_authenticated_session()

    box1_id = create_box_and_post_some_events_to_it(session=s1, close=False)

    print('Testing file upload')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/encrypted-files',
        files={
            'encrypted_file': os.urandom(64),
            'msg_encrypted_content': (None, b64encode(os.urandom(32)).decode()),
            'msg_public_key': (None, b64encode(os.urandom(32)).decode()),
        },
        expected_status_code=201,
    )

    print(f'Testing box closing')
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

    # Testing posting event on unexisting box id
    r = s1.post(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=404
    )

    print('Testing not-found is returned while getting a box with a non-existing id')
    r = s1.get(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
        expected_status_code=404,
    )

    print('Testing non-uuid box in path')
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

    print('Testing incorrect event content format')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.text',
        },
        expected_status_code=400
    )

    print('Testing to create event on closed box is impossible')
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

    print('Testing non-creator canot list events on a closed box')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        expected_status_code=403,
    )

    # Another identity creates other boxes
    box2_id = create_box_and_post_some_events_to_it(session=s2, close=False)
    box3_id = create_box_and_post_some_events_to_it(session=s2)
    box4_id = create_box_and_post_some_events_to_it(session=s2)

    print('Testing identity 1 (non-creator) can list all events on open box')
    r = s1.get(f'{URL_PREFIX}/boxes/{box1_id}/events')
    assert len(r.json()) == 4

    print('Testing identity 1 (non-creator) posts to box 2 a legit event')
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

    print('Testing identity 1 (non-creator) posts to box 2 a creator restricted event')
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

    print('Testing boxes listing')
    r = s1.get(f'{URL_PREFIX}/boxes')
    boxes = r.json()
    assert len(boxes) == 2
    # identity one did not take part into box 3 so it should not be returned
    assert set(map(lambda box: box['id'], boxes)) == {box1_id, box2_id}

    print('Testing pagination')
    r = s2.get(
        f'{URL_PREFIX}/boxes',
        params={
            'offset': 1,
            'limit': 2,
        }
    )
    boxes = r.json()
    assert len(boxes) == 2

    r = s2.get(
        f'{URL_PREFIX}/boxes',
        params={
            'offset': 1,
            'limit': 10,
        }
    )
    boxes = r.json()
    # Identity 2 has 3 boxes in total
    assert len(boxes) == 2

    r = s2.get(
        f'{URL_PREFIX}/boxes',
        params={
            'offset': 20,
            'limit': 2,
        }
    )

    boxes = r.json()
    assert boxes == []

    # Message Edition & Deletion
    # Note that for now message edition and deletion return a HTTP 201 Created
    # but this is likely to change soon

    r = s1.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        }
    )
    box5_id = r.json()['id']

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=201
    )
    text_msg_id = r.json()['id']

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/encrypted-files',
        files={
            'encrypted_file': os.urandom(64),
            'msg_encrypted_content': (None, b64encode(os.urandom(32)).decode()),
            'msg_public_key': (None, b64encode(os.urandom(32)).decode()),
        },
    )
    file_msg_id = r.json()['id']
    encrypted_file_id = r.json()['content']['encrypted_file_id']

    print('Testing message edition')
    s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                'event_id': text_msg_id,
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=201,
    )

    box5_events = s1.get(f'{URL_PREFIX}/boxes/{box5_id}/events').json()

    assert box5_events[1]['content']['encrypted'].startswith("Edited")
    assert box5_events[1]['content']['public_key'].startswith("Edited")
    assert box5_events[1]['content']['last_edited_at']

    print('Testing cannot edit message of type not "msg.text"')
    s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                # Oldest event (last in the list) is the box creation event
                'event_id': box5_events[-1]['id'],
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=401,
    )

    print('Testing cannot edit message of type "msg.file"')
    s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                'event_id': file_msg_id,
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=401,
    )

    print('Testing user cannot edit message she does not own')
    s2.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                'event_id': text_msg_id,
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=401,
    )

    print("Testing (non-admin) user cannot delete message she doesn't own")
    # message is owned by s1, not s2
    s2.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': text_msg_id,
            }
        },
        expected_status_code=401,
    )

    print('Testing "create"-type events cannot be deleted')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                # oldest event is last in the list
                'event_id': box5_events[-1]['id'],
            }
        },
        expected_status_code=403,
    )

    print('Testing deletion of text messages')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': text_msg_id,
            }
        },
        expected_status_code=201,
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['content']['deleted']['by_identity']['identifier']['value'] == s1.email)
        ]
    )

    print('Testing deletion of file messages')
    all_encrypted_files = list_encrypted_files()
    assert encrypted_file_id in all_encrypted_files

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': file_msg_id,
            }
        },
        expected_status_code=201,
    )

    all_encrypted_files = list_encrypted_files()
    assert encrypted_file_id not in all_encrypted_files


    print('Testing cannot delete a message twice')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': text_msg_id,
            }
        },
        expected_status_code=410,
    )

    print('Testing cannot edit a deleted message')
    s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                'event_id': text_msg_id,
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=410,
    )

    print('Testing box admin can delete any message')
    # Message posted by s2 but deleted by s1 (box creator)
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        },
    )
    msg_id = r.json()['id']

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box5_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': msg_id,
            }
        },
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['content']['deleted']['by_identity']['identifier']['value'] == s1.email)
        ]
    )


    print('All OK')
