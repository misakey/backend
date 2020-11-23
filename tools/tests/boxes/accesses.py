#!/usr/bin/env python3

from misapy import URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it, create_add_invitation_link_event
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    box_id, _ = create_box_and_post_some_events_to_it(session=s1, close=False)
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

    print('- identity 2 has been notified about the kick event')
    r = s2.get(
        f'{URL_PREFIX}/identities/{s2.identity_id}/notifications',
        expected_status_code=200
    )
    assert r.json()[0]['details']['id'] == box_id
    assert r.json()[0]['type'] == 'member.kick'

    print('- identity 1 create acceses and identity 2 cannot access')
    invitation_link_event = create_add_invitation_link_event()
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [ invitation_link_event ]
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
    assert r.json()[1]['content'] == invitation_link_event['content']

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

