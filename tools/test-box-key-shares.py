#!/usr/bin/env python3
import os
from base64 import b64encode, urlsafe_b64encode

from misapy.get_access_token import get_authenticated_session
from misapy.test_context import testContext
from misapy.check_response import check_response, assert_fn

with testContext('Box key shares'):
    box_key_share = {
        'share': b64encode(os.urandom(16)).decode(),
        'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
        'box_id': 'df53b67d-5619-4ed5-8e6e-01ea7590b5b0',
    }

    s = get_authenticated_session(require_account=True)

    s.post(
        'https://api.misakey.com.local/box-key-shares',
        json=box_key_share,
    )

    r = s.get(
        f'https://api.misakey.com.local/box-key-shares/{box_key_share["other_share_hash"]}'
    )

    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == box_key_share)
        ]
    )
