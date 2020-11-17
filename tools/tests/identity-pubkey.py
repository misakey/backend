#!/usr/bin/env python3
import os

from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext
from misapy.check_response import check_response, assert_fn
from misapy.utils.base64 import b64encode

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)

    print('- initial state')
    r = s1.get_identity()
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == None),
        ]
    )

    print('- setting pubkey')
    pubkey = b64encode(os.urandom(16))
    s1.set_identity_pubkey(pubkey)

    r = s1.get_identity()
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['pubkey'] == pubkey),
        ]
    )

    print('- retrieving pubkey via identifier')
    r = s1.get_identity_pubkeys(s1.email)
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == [pubkey]),
        ]
    )
