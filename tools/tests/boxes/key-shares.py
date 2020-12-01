#!/usr/bin/env python3
from misapy import http, URL_PREFIX
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext
from misapy.boxes.key_shares import new_key_share_event
from misapy.box_helpers import create_add_invitation_link_event

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)
    # we create an ACR1 member of the box
    # mainly to test cryptoactions (ACR1 members cannot receive them)
    s3 = get_authenticated_session(acr_values=1)

    print('- providing key share during box creation')
    initial_key_share_data = new_key_share_event()['for_server_no_store']
    r = s1.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
            'key_share': initial_key_share_data,
        },
    )
    box_id = r.json()['id']

    other_share_hash = initial_key_share_data['other_share_hash']
    r = s1.get(f'{URL_PREFIX}/box-key-shares/{other_share_hash}')
    check_response(
        r,
        [
            lambda r: r.json()['share'] == initial_key_share_data['misakey_share'],
            lambda r: r.json()['other_share_hash'] == initial_key_share_data['other_share_hash'],
            lambda r: r.json()['box_id'] == box_id,
        ]
    )

    # make box joinable
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [create_add_invitation_link_event()]
        },
        expected_status_code=http.STATUS_CREATED,
    )

    s2.join_box(box_id)
    s3.join_box(box_id)

    print('- get encrypted invitation key share')
    r = s1.get(
        f'{URL_PREFIX}/box-key-shares/encrypted-invitation-key-share',
        params={ 'box_id': box_id },
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == initial_key_share_data['encrypted_invitation_key_share']),
        ]
    )

    print('- update box key share')
    key_share_event = new_key_share_event()
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json=key_share_event,
        expected_status_code=http.STATUS_CREATED,
    )

    other_share_hash = key_share_event['for_server_no_store']['other_share_hash']
    r = s1.get(f'{URL_PREFIX}/box-key-shares/{other_share_hash}')
    check_response(
        r,
        [
            lambda r: r.json()['share'] == key_share_event['for_server_no_store']['misakey_share'],
            lambda r: r.json()['other_share_hash'] == key_share_event['for_server_no_store']['other_share_hash'],
            lambda r: r.json()['box_id'] == box_id,
        ]
    )

    old_other_share_hash = initial_key_share_data['other_share_hash']
    s1.get(
        f'{URL_PREFIX}/box-key-shares/{old_other_share_hash}',
        expected_status_code=http.STATUS_NOT_FOUND,
    )

    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 0),
        ]
    )

    r = s2.get(f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 1),
            lambda r: assert_fn(r.json()[0]['type'] == 'set_box_key_share'),
            lambda r: assert_fn(r.json()[0]['encrypted'] == key_share_event['for_server_no_store']['encrypted_invitation_key_share'])
        ]
    )

    print('- cannot update box key share if not box admin')
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json=new_key_share_event(),
        expected_status_code=http.STATUS_FORBIDDEN,
    )
