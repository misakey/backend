#!/usr/bin/env python3

import os
from base64 import b64encode

from misapy import URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    print('- s1 creates a box')
    box_id, _ = create_box_and_post_some_events_to_it(session=s1, public=False)
    print('- s1 lists members: one member')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200
    )
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 1),
        ]
    )

    print('- s1 makes box public')
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'state.access_mode',
            'content': {
                'value': 'public'
            }
        },
        expected_status_code=201
    )

    print('- s2 join the box')
    join_box(s2, box_id)

    print('- s2 list members: two members - no identifier value')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200
    )
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 2),
        ]
    )
    for member in r.json():
        assert member['identifier_value'] == ""
    
    print('- s1 list members: two members - identifier value')
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200
    )
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 2),
        ]
    )
    for member in r.json():
        assert member['identifier_value'] != ""