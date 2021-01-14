#!/usr/bin/env python3
import json
import os
import sys
import random
from time import sleep
from base64 import b64encode, b64decode

from misapy import http, URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    print('- check identity notifications routes')
    s1 = get_authenticated_session(acr_values=2)

    print('- count notifications of a freshly created user')
    r = s1.head(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications',
        expected_status_code=204
    )
    count_notifs = int(r.headers['X-Total-Count'])

    print('- get notifications and check their content')
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications?offset=0',
        expected_status_code=200
    )
    notif_id = r.json()[0]['id']
    assert r.json()[0]['type'] == 'user.create_account'
    assert r.json()[0]['details'] == None
    assert r.json()[0]['acknowledged_at'] == None
    assert r.json()[1]['type'] == 'user.create_identity'

    print('- acknowldege random id which is not one of the existing ones - 204 is returned because of idempotency')
    id = random.randrange(10000, 100000)
    r = s1.put(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications/acknowledgement?ids={id}',
        expected_status_code=204
    )
    print('- all notifs are still there then')
    r = s1.head(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications',
        expected_status_code=204
    )
    assert int(r.headers['X-Total-Count']) == count_notifs
    # acknowledge for real the notif
    r = s1.put(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications/acknowledgement?ids={notif_id}',
        expected_status_code=204
    )
    print('- the notif is not here anymore')
    r = s1.head(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications',
        expected_status_code=204
    )
    assert int(r.headers['X-Total-Count']) == count_notifs-1
    # notif acknowledged
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/notifications?offset=0&limit=2',
        expected_status_code=200
    )
    assert r.json()[0]['id'] == notif_id
    assert r.json()[0]['type'] == 'user.create_account'
    assert r.json()[0]['details'] == None
    assert r.json()[0]['acknowledged_at'] != None


    print('- check profile routes')
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    # check default email config is private
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        expected_status_code=200
    )
    assert r.json()['email'] == False

   # check the private config is applied to profile
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile',
        expected_status_code=200
    )
    assert r.json()['identifier_value'] == ""
    assert r.json()['identifier_kind'] == ""
    assert r.json()['display_name'] == s1.display_name

    # change the profile config to share email
    s1.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        json={'email': True},
        expected_status_code=204
    )
   # check the sharing is applied in profile
    r = s2.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile',
        expected_status_code=200
    )
    assert r.json()['identifier_value'] == s1.email
    assert r.json()['display_name'] == s1.display_name

   # check the sharing is applied in config
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        expected_status_code=200
    )
    assert r.json()['email'] == True

    # another connected identities cannot change another identities profile config
    s2.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        expected_status_code=403
    )
    s2.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        json={'email': False},
        expected_status_code=403
    )

    # remove share of email for real
    s1.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        json={'email': False},
        expected_status_code=204
    )
   # check the private config is applied back to profile
    r = s2.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile',
        expected_status_code=200
    )
    assert r.json()['identifier_value'] == ''
    assert r.json()['display_name'] == s1.display_name

    # check the config private is applied in config
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        expected_status_code=200
    )
    assert r.json()['email'] == False
