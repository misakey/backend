#!/usr/bin/env python3

from misapy import http, URL_PREFIX
from misapy.check_response import assert_fn, check_response
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)

    print('- new identities have aes-rsa pubkeys')
    r = s1.get(f'{URL_PREFIX}/identities/{s1.identity_id}')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey_aes_rsa'].startswith('com.misakey.aes-rsa-enc:')),
            lambda r: assert_fn(r.json()['non_identified_pubkey_aes_rsa'].startswith('com.misakey.aes-rsa-enc:')),
        ]
    )

    print('- cannot set an identity pubkey with non-default algorithm')
    s1.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}',
        json={
            'pubkey': 'com.misakey.aes-rsa-enc:ShouldBeUnpaddedUrlSafeBase64'
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )
    s1.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}',
        json={
            'non_identified_pubkey': 'com.misakey.aes-rsa-enc:ShouldBeUnpaddedUrlSafeBase64'
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

    print('- cannot set an AES-RSA identity pubkey with bad algorithm')
    s1.patch(
        f'{URL_PREFIX}/identities/{s1.identity_id}',
        json={
            'pubkey': 'com.misakey.BAD-enc:ShouldBeUnpaddedUrlSafeBase64'
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )
