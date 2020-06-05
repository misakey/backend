import json
import os
from time import sleep
from base64 import b64encode

from .. import http

URL_PREFIX = 'http://127.0.0.1:5020'

r = http.post(
    f'{URL_PREFIX}/boxes',
    json={
        'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
        'title': 'Test Box',
    }
)

box_id = r.json()['id']

r = http.post(
    f'{URL_PREFIX}/boxes/{box_id}/events',
    json={
        'type': 'msg.text',
        'content': {
            'encrypted': b64encode(os.urandom(32)).decode()
        }
    }
)

r = http.post(
    f'{URL_PREFIX}/boxes/{box_id}/events',
    json={
        'type': 'msg.file',
        'content': {
            'encrypted': b64encode(os.urandom(32)).decode()
        }
    }
)

r = http.post(
    f'{URL_PREFIX}/boxes/{box_id}/events',
    json={
        'type': 'state.lifecycle',
        'content': {
            'state': 'closed'
        }
    }
)

r = http.get(f'{URL_PREFIX}/boxes/{box_id}/events')
event_list = r.json()
assert len(event_list) == 4
for event in event_list:
    assert 'id' in event
    assert 'server_event_created_at' in event
    assert 'type' in event
    assert 'content' in event

# Testing bad box ID
r = http.post(
    f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
    json={
        'type': 'msg.text',
        'content': {
            'encrypted': b64encode(os.urandom(32)).decode()
        }
    },
    expected_status_code=404
)
r = http.get(
    f'{URL_PREFIX}/boxes/457d5c70-03c2-4179-92a5-f945e666b922/events',
    expected_status_code=404,
)

r = http.post(
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

r = http.post(
    f'{URL_PREFIX}/boxes/{box_id}/events',
    json={
        'type': 'msg.text',
    },
    expected_status_code=400
)

print('All OK')