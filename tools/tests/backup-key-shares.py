#!/usr/bin/env python3
import os
from base64 import b64encode, urlsafe_b64encode

from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext
from misapy.check_response import check_response, assert_fn

with prettyErrorContext():
    print('- normal scenario')

    s = get_authenticated_session(require_account=True)

    backup_key_share = {
        'share': b64encode(os.urandom(16)).decode(),
        'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
        'account_id': s.account_id,
        'salt_base64': b64encode(os.urandom(16)).decode(),
    }

    r = s.post(
        'https://api.misakey.com.local/backup-key-shares',
        json=backup_key_share,
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == backup_key_share)
        ]
    )

    r = s.get(
        f'https://api.misakey.com.local/backup-key-shares/{backup_key_share["other_share_hash"]}'
    )
    check_response(
        r,
        [
            lambda r: assert_fn(r.json() == backup_key_share)
        ]
    )

    print('- Non-existing share')
    s.get(
        f'https://api.misakey.com.local/backup-key-shares/rOdGA-UXBfzNcHqscSfnNQQ',
        expected_status_code=404,
    )

    print('- "Not Found" if share belongs to another account')
    s2 = get_authenticated_session(require_account=True)
    s2.get(
        f'https://api.misakey.com.local/backup-key-shares/{backup_key_share["other_share_hash"]}',
        expected_status_code=404,
    )

    print('- Querier does not have account')
    s3 = get_authenticated_session()

    s3.post(
        'https://api.misakey.com.local/backup-key-shares',
        json=backup_key_share,
        expected_status_code=403,
    )

    r = s3.get(
        f'https://api.misakey.com.local/backup-key-shares/{backup_key_share["other_share_hash"]}',
        expected_status_code=403,
    )


    print('- Forbidden if posting share with bad account ID')
    s.post(
        'https://api.misakey.com.local/backup-key-shares',
        json={
            'share': b64encode(os.urandom(16)).decode(),
            'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
            'account_id': 'df53b67d-5619-4ed5-8e6e-01ea7590b5b0',
            'salt_base64': b64encode(os.urandom(16)).decode(),
        },
        expected_status_code=403,
    )

    print('- Bad request if missing attributes')
    s.post(
        'https://api.misakey.com.local/backup-key-shares',
        json={
            'share': b64encode(os.urandom(16)).decode(),
            'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
        },
        expected_status_code=400,
    )

    s.post(
        'https://api.misakey.com.local/backup-key-shares',
        json={
            'other_share_hash': urlsafe_b64encode(os.urandom(16)).decode().rstrip('='),
            'account_id': s.account_id,
        },
        expected_status_code=400,
    )

    s.post(
        'https://api.misakey.com.local/backup-key-shares',
        json={
            'share': b64encode(os.urandom(16)).decode(),
            'account_id': s.account_id,
        },
        expected_status_code=400,
    )