#!/usr/bin/env python3

import os
from base64 import b64encode

from misapy import URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.box_members import join_box, leave_box
from misapy.check_response import check_response, assert_fn
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    box_id, _ = create_box_and_post_some_events_to_it(session=s1, public=False)
    box_id_2, _ = create_box_and_post_some_events_to_it(session=s1)
    print(f' ~ testing accesses on {box_id}')

    print(f'- identity 1 boxes listing should return {box_id} and {box_id_2}')
    r = s1.get(f'{URL_PREFIX}/boxes/joined', expected_status_code=200)
    boxes = r.json()
    assert len(boxes) == 2
    assert set(map(lambda box: box['id'], boxes)) == {box_id, box_id_2}

    print('- identity 2 is not a member, they cannot get the box')
    s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403
    )

    print("- identity 2 cannot yet become a member of the box")
    s2.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'member.join',
        },
        expected_status_code=403,
    )

    print('- identity 1 can switch the access mode to public')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )
    assert r.json()['access_mode'] == 'limited'
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
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )
    assert r.json()['access_mode'] == 'public'

    print("- identity 2 get a not_member error trying to get the box")
    # identity usually first get the box and have a no_member error before joining the box
    r = s2.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=403,
    )
    assert r.json()['details']['reason'] == 'not_member'

    print("- identity 2 can become a member of the box and is added to the access list")
    join_box(s2, box_id)
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200,
    )
    assert len(r.json()) == 1
    assert r.json()[0]['content']['restriction_type'] == 'identifier'
    assert r.json()[0]['content']['value'] == s2.email
    s2_access_event_id = r.json()[0]['id']

    print('- identity 2 (non-creator) cannot post to box an admin-restricted event')
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'state.access_mode',
            'content': {
                'value': 'limited'
            }
        },
        expected_status_code=403
    )

    print('- identity 1 (creator) can kick in public mode the identity 2')
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.rm', 'referrer_id': s2_access_event_id }
            ]
        },
        expected_status_code=201,
    )

    print('- identity 1 (creator) can add the access to the user can joined in limited mode')
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email
                    }
                }
            ]
        },
        expected_status_code=201,
    )
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'state.access_mode',
            'content': {
                'value': 'limited'
            }
        },
        expected_status_code=201
    )
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200
    )
    assert r.json()['access_mode'] == 'limited'

    print('- identity 2 (non-creator) is not added twice to the access list if already inside it')
    join_box(s2, box_id)
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/accesses',
        expected_status_code=200,
    )
    assert len(r.json()) == 1
    s2_access_event_id = r.json()[0]['id']

    print('- identity 2 (non-creator) can list all events on box')
    r = s2.get(f'{URL_PREFIX}/boxes/{box_id}/events')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 6)
        ]
    )

    print('- identity 2 (non-creator) posts to box a legit event')
    r = s2.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'msg.text',
            'content': {
                'encrypted': b64encode(os.urandom(32)).decode(),
                'public_key': b64encode(os.urandom(32)).decode()
            }
        }
    )

    print(f'- identity 2 boxes listing should return box {box_id}')
    r = s2.get(f'{URL_PREFIX}/boxes/joined', expected_status_code=200)
    boxes = r.json()
    assert len(boxes) == 1
    assert boxes[0]['id'] == box_id
    
    print('- identity 1 can still kick the identity 2 in limited mode')
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                { 'type': 'access.rm', 'referrer_id': s2_access_event_id }
            ]
        },
        expected_status_code=201
    )
    
    print('- identity 1 have an empty list of accesses afterward')
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

    print('- identity 2 has been notified about the kick event')
    r = s2.get(
        f'{URL_PREFIX}/identities/{s2.identity_id}/notifications',
        expected_status_code=200
    )
    assert r.json()[0]['details']['id'] == box_id
    assert r.json()[0]['type'] == 'member.kick'

    print('- identity 1 create accesses and identity 2 cannot access because it does not comply')
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
    assert len(r.json()) == 2
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

    print('- identity 2 cannot retrieve the box anymore')
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
    assert len(r.json()) == 1
    assert r.json()[0]['content'] == access_email_domain

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
                { 'type': 'access.add'}, { 'type': 'state.access_mode'},
            ]
        },
        expected_status_code=400,
    )
    assert r.json()['details']['events'] != ''

