#!/usr/bin/env python3
import os

from misapy.test_context import testContext
from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext
from misapy.check_response import check_response, assert_fn
from misapy.utils.base64 import b64encode

with testContext():
    s1 = get_authenticated_session(acr_values=2)

with testContext('initial state'):
    r = s1.get_identity()
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == None),
        ]
    )

with testContext('setting pubkey'):
    pubkey = b64encode(os.urandom(16))
    s1.set_identity_pubkey(pubkey)

    r = s1.get_identity()
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == pubkey),
        ]
    )

with testContext('retrieving pubkey via identifier'):
    r = s1.get_identity_pubkeys(s1.email)
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == [pubkey]),
        ]
    )
