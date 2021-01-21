#!/usr/bin/env python3

import json
import os
import sys
from time import sleep

from . import http, URL_PREFIX
from .utils.base64 import b64encode, urlsafe_b64encode
from .get_access_token import get_authenticated_session
from .container_access import list_encrypted_files
from .check_response import check_response, assert_fn
from .boxes.key_shares import new_key_share_event

def create_box_and_post_some_events_to_it(session, public=True):
    s = session

    r = s.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        }
    )
    creator = r.json()['creator']
    assert creator['identifier_value'] == s.email

    box_id = r.json()['id']

    print('- set box key share')
    key_share_event = new_key_share_event()
    s.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json=key_share_event,
        expected_status_code=http.STATUS_CREATED,
    )
    other_share_hash = key_share_event['extra']['other_share_hash']

    print(f'- create msg.text event on box {box_id}')
    r = s.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)),
                'public_key': b64encode(os.urandom(32)),
            }
        },
        expected_status_code=201
    )

    expected_event_count = 3

    if public:
        expected_event_count += 1
        print(f'- set access mode to public for box {box_id}')
        s.post(
            f'{URL_PREFIX}/boxes/{box_id}/events',
            json={
                'type': 'state.access_mode',
                'content': {
                    'value': 'public',
                }
            },
            expected_status_code=201,
        )

    print(f'- listing for created box {box_id}')
    r = s.get(f'{URL_PREFIX}/boxes/{box_id}/events')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == expected_event_count),
        ]
    )
    for event in r.json():
        assert 'id' in event
        assert 'server_event_created_at' in event
        assert 'type' in event
        assert 'content' in event
        # impossible to see identifier value while listing events
        assert event['sender']['identifier_value'] == ""

    return box_id, other_share_hash
