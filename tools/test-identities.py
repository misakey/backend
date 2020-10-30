#!/usr/bin/env python3
import json
import os
import sys
from time import sleep
from base64 import b64encode, b64decode

from misapy import http, URL_PREFIX
from misapy.box_helpers import create_box_and_post_some_events_to_it, create_add_invitation_link_event
from misapy.container_access import list_encrypted_files
from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext


with testContext('check profile routes'):
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
    assert r.json()['identifier_id'] == ""
    assert r.json()['identifier']['value'] == ""
    assert r.json()['identifier']['kind'] == ""
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
    assert r.json()['identifier']['value'] == s1.email
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
    assert r.json()['identifier']['value'] == ''
    assert r.json()['display_name'] == s1.display_name

    # check the config private is applied in config
    r = s1.get(
        f'{URL_PREFIX}/identities/{s1.identity_id}/profile/config',
        expected_status_code=200
    )
    assert r.json()['email'] == False
