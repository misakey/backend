#!/usr/bin/env python3

import json

from . import URL_PREFIX
from .check_response import check_response,assert_fn

def create_org(session):
    s = session

    r = s.post(
        f'{URL_PREFIX}/organizations',
        json={
            'name': 'AwesomeOrg',
        },
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['name'] == 'AwesomeOrg'),
            lambda r: assert_fn(r.json()['creator_id'] == s.identity_id),
            lambda r: assert_fn(r.json()['current_identity_role'] == 'admin'),
        ]
    )

    return r.json()['id']

def create_datatag(session, org_id):
    s = session

    r = s.post(
        f'{URL_PREFIX}/organizations/{org_id}/datatags',
        json={
            'name': 'AwesomeDatatag',
        },
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json()['name'] == 'AwesomeDatatag'),
            lambda r: assert_fn(r.json()['organization_id'] == org_id),
        ]
    )

    return r.json()['id']
