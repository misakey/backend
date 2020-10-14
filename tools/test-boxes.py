#!/usr/bin/env python3

import json
import os
import sys
from time import sleep
from base64 import b64encode, b64decode

from misapy import http
from misapy.box_helpers import URL_PREFIX, create_box_and_post_some_events_to_it
from misapy.box_key_shares import create_key_share, get_key_share
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext

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
    box2_id, _ = create_box_and_post_some_events_to_it(session=s2, close=False)
    create_box_and_post_some_events_to_it(session=s2)
    create_box_and_post_some_events_to_it(session=s2)    

    print("- identity 1 becomes member of box 2")
    join_box(s1, box2_id)

    print('- identity 1 (creator) can list all events on open box 2')
    r = s1.get(f'{URL_PREFIX}/boxes/{box2_id}/events')
    assert len(r.json()) == 3

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

def test_box_messages(s1, s2): 
    # Message Edition & Deletion
    box1_id, _ = create_box_and_post_some_events_to_it(session=s1, close=False)
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
            'referrer_id': text_msg_id,
            'content': {
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
            # Oldest event (last in the list) is the box creation event
            'referrer_id': box5_events[-1]['id'],
            'content': {
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=403,
    )

    print('- cannot edit message of type "msg.file"')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.edit',
            'referrer_id': file_msg_id,
            'content': {
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=403,
    )

    print('- user cannot edit message they do not own')
    s2.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.edit',
            'referrer_id': text_msg_id,
            'content': {
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
            'referrer_id': text_msg_id
        },
        expected_status_code=403,
    )

    print('- "create"-type events cannot be deleted')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            # oldest event is last in the list
            'referrer_id': box5_events[-1]['id']
        },
        expected_status_code=403,
    )

    print('- deletion of text messages')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            'referrer_id': text_msg_id
        },
        expected_status_code=201,
    )
    assert r.json()['referrer_id'] == text_msg_id
    assert r.json()['sender']['identifier_id'] == s1.identifier_id

    print(f'- deletion of file message {file_msg_id}')
    all_encrypted_files = list_encrypted_files()
    assert encrypted_file_id in all_encrypted_files

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.delete',
            'referrer_id': file_msg_id
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
            'referrer_id': text_msg_id
        },
        expected_status_code=410,
    )

    print('- cannot edit a deleted message')
    s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.edit',
            'referrer_id': text_msg_id,
            'content': {
                'new_encrypted': b64encode(b64decode('EditedXX') + os.urandom(32)).decode(),
                'new_public_key': b64encode(b64decode('EditedXX') + os.urandom(32)).decode()
            }
        },
        expected_status_code=410,
    )

    print('- box admin can delete any message')
    # Message posted by s2 but deleted by s1 (box creator) - become first a member
    join_box(s2, box1_id)
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
            'referrer_id': msg_id
        },
    )
    assert r.json()['referrer_id'] == msg_id
    assert r.json()['sender']['identifier_id'] == s1.identifier_id

def test_accesses(s1, s2):
    box_id, box_share_hash = create_box_and_post_some_events_to_it(session=s1, close=False)
    print(f' ~ testing accesses on {box_id}')

    print('- identity 2 is not a member, they cannot get the box')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 2 becomes a member by getting the key share')
    join_box(s2, box_id)

    print('- identity 2 can then can get the box and see its creator')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )
    assert r.json()['creator']['identifier_id'] == s1.identifier_id


    print('- identity 1 makes the box now private')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    referrer_id = r.json()[0]["id"]
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.rm', 'referrer_id': referrer_id }
            ]
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
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.add', 'content': access_invitation_link }
            ]
        },
        expected_status_code=201,
    )
    access_identifier = {
        'restriction_type': 'identifier',
        'value': 'email@random.io'  
    }
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.add', 'content': access_identifier}
            ]
        },
        expected_status_code=201,
    )
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- correct identifier rule gives access to identity 2')
    access_identifier['value'] = s2.email
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.add', 'content': access_identifier }
            ]
        },
        expected_status_code=201,
    )
    join_id = join_box(s2, box_id).json()['id']

    # retrieve accesses to remove the recently added access
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    assert len(r.json()) == 3
    assert r.json()[0]['content']['restriction_type'] == 'identifier'
    assert r.json()[0]['content']['value'] == s2.email
    s2_email_referrer_id = r.json()[0]['id']
    random_email_referrer_id = r.json()[1]['id']

    print('- identity 2 can list members but cannot see identifier.value')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200
    )
    for member in r.json():
        assert member['identifier']['value'] == ""

    print('- batch events removes accesses and add email_domain')
    access_email_domain = {
        'restriction_type': 'email_domain',
        'value': 'company.com'
    }
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.rm', 'referrer_id': s2_email_referrer_id},
                { 'type': 'access.rm', 'referrer_id': random_email_referrer_id},
                { 'type': 'access.add', 'content': access_email_domain}
            ]
        },
        expected_status_code=201,
    )
    # check member.kick event refer to the previously created join event
    assert len(r.json()) == 4
    assert r.json()[3]['type'] == 'member.kick'
    assert r.json()[3]['referrer_id'] == join_id

    print('- identity 2 is kicked and cannot retrieve the box')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print('- identity 2 is kicked and is not part anymore of the members list')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200
    )
    assert len(r.json()) == 1
    assert r.json()[0]['identifier']['value'] == s1.email
    
    print('- the box is not listed when the user requested their boxes')
    r = s2.get(
        f'{URL_PREFIX}/boxes/joined',
        expected_status_code=200
    )
    for box in r.json():
        assert box['id'] != box_id

    print('- the removed access is not listed anymore')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    )
    assert len(r.json()) == 2
    assert r.json()[0]['content'] == access_email_domain
    assert r.json()[1]['content'] == access_invitation_link

    print('- email_domain access rule gives correctly accesses')
    access_email_domain['value'] = 'misakey.com'
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.add', 'content': access_email_domain }
            ]
        },
        expected_status_code=201,
    )
    join_box(s2, box_id)

    print('- fail batch events request do no add/rm any event in the batch')
    accessesBefore = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    ).json()
    access_email_domain['value'] = 'labri.cot'
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.rm', 'referrer_id': accessesBefore[0]['id']},
                { 'type': 'access.add', 'content': access_email_domain},
                # /!\ non existing UUID as referrer_id that will make the DB layer failing
                { 'type': 'access.rm', 'referrer_id': '687a3dbc-8955-46be-a9a7-e7ad0daad74c'},
            ]
        },
        expected_status_code=404,
    )
    accessesAfter = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200
    ).json()
    assert accessesBefore == accessesAfter


    print('- batch type format is checked and events are required')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'lol',
            'events' : []
        },
        expected_status_code=400,
    )
    assert r.json()['details']['batch_type'] == 'malformed'
    assert r.json()['details']['events'] == 'required'

    print('- batch events type format is checked')
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.add'}, { 'type': 'state.lifecycle'},
            ]
        },
        expected_status_code=400,
    )
    assert r.json()['details']['events'] != ''


with testContext():
    # Init 2 user sessions for creator rules testing
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    print('--------\nBasics...')
    test_basics(s1, s2)

    print('--------\nMessages...')
    test_box_messages(s1, s2)

    print('--------\nAccesses...')
    test_accesses(s1, s2)
    print('All OK')
