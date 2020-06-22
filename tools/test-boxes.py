#!/usr/bin/env python3

import json
import os
import sys
from time import sleep
from base64 import b64encode

from misapy import http
from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext

URL_PREFIX = 'http://127.0.0.1:5020'


def create_box_and_post_some_events_to_it(session):
    s = session

    r = s.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        }
    )
    creator = r.json()['creator']
    assert creator['display_name'] == s.email
    assert creator['identifier']['value'] == s.email

    box_id = r.json()['id']

    r = s.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode()
            }
        }
    )

    r = s.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'state.lifecycle',
            'content': {
                'state': 'closed'
            }
        }
    )

    r = s.get(f'{URL_PREFIX}/boxes/{box_id}/events')
    event_list = r.json()
    assert len(event_list) == 3
    for event in event_list:
        assert 'id' in event
        assert 'server_event_created_at' in event
        assert 'type' in event
        assert 'content' in event

        sender = event['sender']
        assert sender['display_name'] == s.email
        assert sender['identifier']['value'] == s.email

    return box_id


with testContext():
    s1 = get_authenticated_session()
    box1_id = create_box_and_post_some_events_to_it(session=s1)

    # Testing bad box ID
    r = s1.post(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=404
    )
    r = s1.get(
        f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
        expected_status_code=404,
    )

    r = s1.post(
        f'{URL_PREFIX}/boxes/YOU_KNOW_IM_BAD/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode()
            }
        },
        expected_status_code=400
    )

    # Testing incomplete event

    r = s1.post(
        f'{URL_PREFIX}/boxes/{box1_id}/events',
        json={
            'type': 'msg.text',
        },
        expected_status_code=400
    )

    # Testing box listing

    # Another identity creates other boxes
    sleep(0.5)  # TODO disable rate limiting in test environment
    s2 = get_authenticated_session()
    box2_id = create_box_and_post_some_events_to_it(session=s2)
    box3_id = create_box_and_post_some_events_to_it(session=s2)
    box4_id = create_box_and_post_some_events_to_it(session=s2)

    # Identity 1 posts to box 2
    r = s1.post(
        f'{URL_PREFIX}/boxes/{box2_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode()
            }
        }
    )

    r = s1.get(f'{URL_PREFIX}/boxes')
    boxes = r.json()
    assert len(boxes) == 2
    # identity one did not take part into box 3 so it should not be returned
    assert set(map(lambda box: box['id'], boxes)) == {box1_id, box2_id}

    # Testing pagination

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

    print('All OK')
