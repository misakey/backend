#!/usr/bin/env python3

import os
from base64 import b64encode, b64decode

from misapy import URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.box_members import join_box
from misapy.container_access import list_encrypted_files
from misapy.pretty_error import prettyErrorContext
from misapy.get_access_token import get_authenticated_session

with prettyErrorContext(): 
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    # Message Edition & Deletion
    box1_id, _ = create_box_and_post_some_events_to_it(session=s1)
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
    assert r.json()['sender']['id'] == s1.identity_id

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
    assert r.json()['sender']['id'] == s1.identity_id
