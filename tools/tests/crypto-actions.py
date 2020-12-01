#!/usr/bin/env python3

import subprocess
from uuid import uuid4

from misapy import http
from misapy.get_access_token import get_authenticated_session
from misapy.check_response import check_response, assert_fn
from misapy.pretty_error import prettyErrorContext

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

with prettyErrorContext():
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

    print('- Listing one\'s actions')
    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 5)
        ]
    )
    actions = r.json()

    print('- Getting one specific action')
    action = actions[0]
    r = s1.get(f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions/{action["id"]}')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == action)
        ]
    )

    print('- Deleting one specific action')
    s1.delete(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions/{action["id"]}',
        expected_status_code=http.STATUS_NO_CONTENT
    )

    print('- No actions')
    r = s2.get(f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == [])
        ]
    )

    print('- Cannot list someone else\'s actions')
    s2.get(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions',
        expected_status_code=403,
    )

    print("- Cannot get someone else's action")
    # s2 cannot use s1's account ID
    s2.get(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions/{actions[0]["id"]}',
        expected_status_code=http.STATUS_FORBIDDEN,
    )

    print("- Cannot get an action not tied to account in path")
    # this time s2 uses her own account ID but still tries to get s1's action
    s2.get(
        f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions/{actions[0]["id"]}',
        expected_status_code=http.STATUS_NOT_FOUND,
    )

    print("- Cannot delete someone else's action")
    # s2 cannot use s1's account ID
    s2.delete(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions/{actions[0]["id"]}',
        expected_status_code=http.STATUS_FORBIDDEN,
    )

    print("- Cannot delete an action not tied to account in path")
    # this time s2 uses her own account ID but still tries to delete s1's action
    s2.delete(
        f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions/{actions[0]["id"]}',
        expected_status_code=http.STATUS_NOT_FOUND,
    )
