#!/usr/bin/env python3
import os
from base64 import b64encode, urlsafe_b64encode

from . import URL_PREFIX
from .check_response import check_response, assert_fn
from .get_access_token import get_authenticated_session

def join_box(session, box_id):
    return session.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'member.join',
        },
        expected_status_code=201,
    )

def leave_box(session, box_id):
    return session.post(
        f'{URL_PREFIX}/boxes/{box_id}/events',
        json={
            'type': 'member.leave',
        },
        expected_status_code=201,
    )
