#!/usr/bin/env python3

import json
import os
import sys
from time import sleep
from base64 import b64encode, b64decode

from misapy import http
from misapy.get_access_token import get_authenticated_session
from misapy.box_key_shares import create_key_share, get_key_share
from misapy.box_helpers import URL_PREFIX, create_box_and_post_some_events_to_it
from misapy.test_context import testContext
from misapy.container_access import list_encrypted_files
from misapy.check_response import check_response, assert_fn

def test_basics(s1, s2):
    box1_id, box1_share_hash = create_box_and_post_some_events_to_it(session=s1, close=False)    

    print("- box key share retrieval")
    r = get_key_share(s1, box1_share_hash)
    check_response(r,[lambda r: assert_fn(r.json()["other_share_hash"] == box1_share_hash)])

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
    box2_id, box2_share_hash = create_box_and_post_some_events_to_it(session=s2, close=False)
    create_box_and_post_some_events_to_it(session=s2)
    create_box_and_post_some_events_to_it(session=s2)    

    print("- identity 1 becomes member of box 2")
    get_key_share(s1, box2_share_hash)

    print('- identity 1 (creator) can list all events on open box 2')
    r = s1.get(f'{URL_PREFIX}/boxes/{box2_id}/events')
    assert len(r.json()) == 2

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

    print('- boxes listing')
    r = s1.get(f'{URL_PREFIX}/boxes')
    boxes = r.json()
    assert len(boxes) == 2
    
    # identity one did not take part into box 3 so it should not be returned
    assert set(map(lambda box: box['id'], boxes)) == {box1_id, box2_id}

    print('- pagination')
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

def test_box_messages(s1, s2): 
    # Message Edition & Deletion
    box1_id, box1_share_hash = create_box_and_post_some_events_to_it(session=s1, close=False)
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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
        f'{URL_PREFIX}/boxes/{box1_id}/encrypted-files',
        files={
            'encrypted_file': os.urandom(64),
            'msg_encrypted_content': (None, b64encode(os.urandom(32)).decode()),
            'msg_public_key': (None, b64encode(os.urandom(32)).decode()),
        },
    )
    file_msg_id = r.json()['id']
    encrypted_file_id = r.json()['content']['encrypted_file_id']

    print('- message edition')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

    box5_events = s1.get(f'{URL_PREFIX}/boxes/{box1_id}/events').json()

    assert box5_events[1]['content']['encrypted'].startswith("Edited")
    assert box5_events[1]['content']['public_key'].startswith("Edited")
    assert box5_events[1]['content']['last_edited_at']

    print('- cannot edit message of type not "msg.text"')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

    print('- cannot edit message of type "msg.file"')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

    print('- user cannot edit message they do not own')
    s2.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.edit',
            'content': {
                'event_id': text_msg_id,
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=403,
    )

    print("- (non-admin) user cannot delete message they do not own")
    # message is owned by s1, not s2
    s2.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': text_msg_id,
            }
        },
        expected_status_code=403,
    )

    print('- "create"-type events cannot be deleted')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                # oldest event is last in the list
                'event_id': box5_events[-1]['id'],
            }
        },
        expected_status_code=403,
    )

    print('- deletion of text messages')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

    print('- deletion of file messages')
    all_encrypted_files = list_encrypted_files()
    assert encrypted_file_id in all_encrypted_files

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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


    print('- cannot delete a message twice')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            'content': {
                'event_id': text_msg_id,
            }
        },
        expected_status_code=410,
    )

    print('- cannot edit a deleted message')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

    print('- box admin can delete any message')
    # Message posted by s2 but deleted by s1 (box creator) - become first a member
    get_key_share(s2, box1_share_hash)
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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
        f'{URL_PREFIX}/boxes/{box1_id}/events',
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

def test_accesses(s1, s2):
    box_id, box_share_hash = create_box_and_post_some_events_to_it(session=s1, close=False)

    print('- identity 2 is not a member, they cannot get the box')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 2 becomes a member by getting the key share')
    get_key_share(s2, box_share_hash)

    print('- identity 2 can then can get the box')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )

    print('- identity 1 makes the box now private')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    referrer_id = r.json()[0]["id"]
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.rm',
            'referrer_id': referrer_id
        },
        expected_status_code=201,
    )
    
    print('- identity 1 do not see the access anymore by listing them')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    assert len(r.json()) == 0

    print('- identity 2 cannot get the box anymore')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 1 create acceses and identity 2 cannot access')
    access_invitation_link = {
        'restriction_type': 'invitation_link',
        'value': box_share_hash
    }
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.add',
            'content': access_invitation_link
        },
        expected_status_code=201,
    )
    access_any_identifier = {
        'restriction_type': 'identifier',
        'value': 'any_identifier'  
    }
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.add',
            'content': access_any_identifier
        },
        expected_status_code=201,
    )
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 1 create identifier rule complying with identity 2')
    access_identity2_identifier = {
        'restriction_type': 'identifier',
        'value': s2.email
    }
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.add',
            'content': access_identity2_identifier
        },
        expected_status_code=201,
    )
    identifier_2_access = r.json()
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )

    print('- identity 1 removes this access')
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.rm',
            'referrer_id': identifier_2_access['id']
        },
        expected_status_code=201,
    )
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- the remove access is not listed anymore')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    assert len(r.json()) == 2
    assert r.json()[0]['content'] == access_any_identifier
    assert r.json()[1]['content'] == access_invitation_link

    print('- identity 1 create email_domain rule not complying with identity 2')
    access_misakey_email_domain = {
        'restriction_type': 'email_domain',
        'value': 'not_misakey.com'
    }
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.add',
            'content': access_misakey_email_domain
        },
        expected_status_code=201,
    )
    identifier_2_access = r.json()
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 1 create email_domain rule matching identity 2 email domain')
    access_misakey_email_domain = {
        'restriction_type': 'email_domain',
        'value': 'misakey.com'
    }
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'access.add',
            'content': access_misakey_email_domain
        },
        expected_status_code=201,
    )
    identifier_2_access = r.json()
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )

with testContext():
    # Init 2 user sessions for creator rules testing
    s1 = get_authenticated_session()
    s2 = get_authenticated_session()

    print('--------\nBasics...')
    test_basics(s1, s2)
    print('--------\nMessages...')
    test_box_messages(s1, s2)
    print('--------\nAccesses...')
    test_accesses(s1, s2)
    print('All OK')
