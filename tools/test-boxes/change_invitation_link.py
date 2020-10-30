#!/usr/bin/env python3
from misapy import http, URL_PREFIX
from misapy.box_helpers import create_add_invitation_link_event
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.box_members import join_box
from misapy.test_context import testContext

with testContext('init invitation changing link test'):
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    # create a box
    r = s1.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        },
    )
    box_id = r.json()['id']

    # a invitation link access to it
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [create_add_invitation_link_event()],
        },
    )
    # retrieve the corresponding access id and join the box
    r = s1.get(f'{URL_PREFIX}/boxes/{box_id}/accesses')
    current_invitation_link_event = r.json()[0]
    join_box(s2, box_id)

with testContext('cannot reset invitation link if one is already active'):
    new_invitation_link_event = create_add_invitation_link_event()
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events': [
                new_invitation_link_event
            ]
        },
        expected_status_code=http.STATUS_CONFLICT,
    )

with testContext('changing invitation link'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events': [
                {
                    'type': 'access.rm',
                    'referrer_id': current_invitation_link_event['id'],

                },
                new_invitation_link_event
            ]
        }
    )

    previous_invitation_link_event = current_invitation_link_event
    del current_invitation_link_event

    # check the previous request has removed the old invitation link and created the new one
    r = s1.get(f'{URL_PREFIX}/boxes/{box_id}/accesses')
    current_invitation_links = [
        event
        for event in r.json()
        if event['content']['restriction_type'] == 'invitation_link'
    ]
    assert len(current_invitation_links) == 1
    assert current_invitation_links[0]['content'] == new_invitation_link_event['content']

    # previous other share doesn't exist anymore
    s1.get(
        f'{URL_PREFIX}/box-key-shares/'+previous_invitation_link_event['content']['value'],
        expected_status_code=http.STATUS_NOT_FOUND,
    )
    # new one does exist
    s1.get(
        f'{URL_PREFIX}/box-key-shares/'+new_invitation_link_event['content']['value'],
        expected_status_code=http.STATUS_OK,
    )
    
    # s2 must not have been kicked out
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=http.STATUS_OK,
    )
    
    # crypto actions must have been created for other members
    r = s2.get(f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 1),
            lambda resp: assert_fn(resp.json()[0]['box_id'] == box_id),
            lambda resp: assert_fn(resp.json()[0]['type'] == 'set_box_key_share')
        ]
    )


