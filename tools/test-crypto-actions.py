#!/usr/bin/env python3

import subprocess
from uuid import uuid4

from misapy import http
from misapy.test_context import testContext
from misapy.get_access_token import get_authenticated_session
from misapy.check_response import check_response, assert_fn

URL_PREFIX = 'https://api.misakey.com.local'

def create_crypto_action(owner, sender, box_id, fake_data):
    sql_command = ' '.join([
        "INSERT INTO crypto_action",
        "(id, account_id, sender_identity_id, type, box_id, encryption_public_key, encrypted)",
        "VALUES",
        f"('{uuid4()}', '{owner.account_id}', '{owner.identity_id}',",
        f"'invitation', '{box_id}', '{fake_data}', '{fake_data}') "
    ])

    proc = subprocess.run(
        (
            'docker exec test-and-run_api_db_1  psql -t -d sso -U misakey -h localhost -c'.split()
            + [sql_command]
        ),
        capture_output=True,
    )
    proc.check_returncode()

with testContext('Listing one\'s actions'):
    s1 = get_authenticated_session(require_account=True)
    s2 = get_authenticated_session(require_account=True)

    r = s2.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        }
    )
    box_id = r.json()['id']

    for i in range(5):
        create_crypto_action(s1, s2, box_id, f'Fake Data Action {i}')

    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 5)
        ]
    )
    actions = r.json()

with testContext('No actions'):
    r = s2.get(f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == [])
        ]
    )

with testContext('Deleting actions'):
    s1.delete(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions',
        json={
            'until_action_id': actions[3]['id'],
        },
        expected_status_code=204,
    )

    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 1),
            lambda r: assert_fn(r.json()[0]['id'] == actions[-1]['id']),
        ]
    )

with testContext('Cannot list someone else\'s actions'):
    s2.get(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions',
        expected_status_code=403,
    )

with testContext('Cannot deleted someone else\'s actions'):
    # action is owned by s1 not s2
    s2.delete(
        f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions',
        json={
            'until_action_id': actions[-1]['id'],
        },
        # We get a "not found" error as if the action does not exist
        expected_status_code=404,
    )

with testContext('"Not found" on deleting non existing action'):
    s1.delete(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions',
        json={
            # shouldn't exist anymore
            'until_action_id': actions[1]['id'],
        },
        expected_status_code=404,
    )
