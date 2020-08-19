#!/usr/bin/env python3
import subprocess
from base64 import b64encode

from misapy.get_access_token import (
    get_authenticated_session,
    get_credentials,
)
from misapy.test_context import testContext

def get_archived_backup(account_id):
    proc = subprocess.run(
        (
            'docker exec test-and-run_api_db_1  psql -t -d sso -U misakey -h localhost -c'.split()
            + [
                "SELECT data "
                "FROM backup_archive "
                f"WHERE account_id = '{account_id}' "
                "ORDER BY created_at DESC LIMIT 1;"
            ]
        ),
        capture_output=True,
    )
    proc.check_returncode()
    output = proc.stdout.decode()
    archived_backup = output.strip()
    return archived_backup

def get_current_backup(account_id):
    proc = subprocess.run(
        (
            'docker exec test-and-run_api_db_1  psql -t -d sso -U misakey -h localhost -c'.split()
            + [
                "SELECT backup_data "
                "FROM account "
                f"WHERE id = '{account_id}' "
            ]
        ),
        capture_output=True,
    )
    proc.check_returncode()
    output = proc.stdout.decode()
    current_backup = output.strip()
    return current_backup

with testContext('archive creation'):
    # We create an account, and then we reset its password
    creds = get_credentials(require_account=True)
    s = get_authenticated_session(email=creds.email, reset_password=True)

    archived_backup = get_archived_backup(creds.account_id)
    assert archived_backup == b64encode(b'fake backup data').decode()

    current_backup = get_current_backup(creds.account_id)
    assert current_backup == b64encode(b'other fake backup data').decode()

    r = s.get('https://api.misakey.com.local/backup-archives')
    assert len(r.json()) == 1

    archive_id = r.json()[0]['id']
    r = s.get(f'https://api.misakey.com.local/backup-archives/{archive_id}/data')
    assert r.json() == b64encode(b'fake backup data').decode()

    # creating some more archives
    s = get_authenticated_session(email=creds.email, reset_password=True)
    s = get_authenticated_session(email=creds.email, reset_password=True)
    r = s.get('https://api.misakey.com.local/backup-archives')
    assert len(r.json()) == 3

    archives = r.json()

with testContext('archive deletion'):
    s.delete(
        f'''https://api.misakey.com.local/backup-archives/{archives[0]['id']}''',
        json={
            'reason': 'deletion',
        }
    )

    s.delete(
        f'''https://api.misakey.com.local/backup-archives/{archives[1]['id']}''',
        json={
            'reason': 'recovery',
        }
    )

    # assuming archives are always returned in the same order
    # (which should be true)
    r = s.get('https://api.misakey.com.local/backup-archives')
    assert len(r.json()) == 3
    # the one we deleted
    assert r.json()[0]['deleted_at'] != None
    assert r.json()[0]['recovered_at'] == None
    # the one we recovered
    assert r.json()[1]['deleted_at'] == None
    assert r.json()[1]['recovered_at'] != None
    # the one we didn't touch
    assert r.json()[2]['deleted_at'] == None
    assert r.json()[2]['recovered_at'] == None

    s.get(
        f'''https://api.misakey.com.local/backup-archives/{archives[0]['id']}/data''',
        expected_status_code=410,
    )

with testContext('bad request on bad deletion reason'):
    s.delete(
        f'''https://api.misakey.com.local/backup-archives/{archives[0]['id']}''',
        json={
            'reason': 'BAAAD',
        },
        expected_status_code=400,
    )

with testContext('cannot re-delete an archive'):
    s.delete(
        f'''https://api.misakey.com.local/backup-archives/{archives[0]['id']}''',
        json={
            'reason': 'deletion',
        },
        expected_status_code=410,
    )

    s.delete(
        f'''https://api.misakey.com.local/backup-archives/{archives[1]['id']}''',
        json={
            'reason': 'recovery',
        },
        expected_status_code=410,
    )

with testContext('archive not found'):
    wrong_id = '74f650f3-b338-40d0-a8c8-37b6c3947522'
    s.get(
        f'https://api.misakey.com.local/backup-archives/{wrong_id}/data',
        expected_status_code=404,
    )

    s.delete(
        f'https://api.misakey.com.local/backup-archives/{wrong_id}',
        json={
            'reason': 'deletion',
        },
        expected_status_code=404,
    )

