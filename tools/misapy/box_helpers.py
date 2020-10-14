#!/usr/bin/env python3

import json
import os
import sys
from time import sleep
from base64 import b64encode, b64decode

from . import http
from .get_access_token import get_authenticated_session
from .box_key_shares import create_key_share, get_key_share
from .test_context import testContext
from .container_access import list_encrypted_files
from .check_response import check_response, assert_fn

URL_PREFIX = 'https://api.misakey.com.local'

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

    print(f'- create msg.text event on box {box_id}')
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

    print(f'- key share creation for box {box_id}')
    key_share = create_key_share(s, box_id).json()

    print(f'- access invitation_link creation for box {box_id}')
    s.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'invitation_link',
                        'value': key_share["other_share_hash"]
                    }
                }
            ]
        },
        expected_status_code=201,
    )

    if close:
        print(f'- close box {box_id}')
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

    print(f'- listing for created box {box_id}')
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

    return box_id, key_share["other_share_hash"]
