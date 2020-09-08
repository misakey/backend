#!/usr/bin/env python3
import os
from base64 import b64encode, urlsafe_b64encode

from .get_access_token import get_authenticated_session
from .test_context import testContext
from .check_response import check_response, assert_fn

def create_key_share(session, box_id):
    box_key_share = {
        'share': b64encode(os.urandom(16)).decode(),
        'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
        'box_id': box_id,
    }
    return session.post(
        'https://api.misakey.com.local/box-key-shares',
        json=box_key_share,
        expected_status_code=201,
    )

def get_key_share(session, other_share_hash):
    return session.get(
        f'https://api.misakey.com.local/box-key-shares/{other_share_hash}'
    )
