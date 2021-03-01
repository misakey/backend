#!/usr/bin/env python3

from misapy import http, URL_PREFIX, SELF_CLIENT_ID
from misapy.pretty_error import prettyErrorContext
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session

with prettyErrorContext():
    s = get_authenticated_session(acr_values=2)


    print('- creating box with alternate encryption algorithm')
    public_key = 'com.misakey.aes-rsa-enc:ShouldBeUnpaddedUrlSafeBase64'
    r = s.post(
        f'{URL_PREFIX}/boxes',
        json={
            'owner_org_id': SELF_CLIENT_ID,
            'public_key': public_key,
            'title': 'Test Box',
        },
        expected_status_code=http.STATUS_CREATED,
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['public_key'] == public_key),
        ]
    )


    print('- "Bad Request" on invalid pubkey algo')
    s.post(
        f'{URL_PREFIX}/boxes',
        json={
            'owner_org_id': SELF_CLIENT_ID,
            'public_key': 'com.misakey.BAD-rsa-enc:ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )