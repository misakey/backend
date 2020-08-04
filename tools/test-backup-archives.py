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

with testContext('Backup Archives'):
    # We create an account, and then we reset its password
    creds = get_credentials(require_account=True)
    s = get_authenticated_session(email=creds.email, reset_password=True)

    archived_backup = get_archived_backup(creds.account_id)
    assert archived_backup == b64encode(b'fake backup data').decode()

    current_backup = get_current_backup(creds.account_id)
    assert current_backup == b64encode(b'other fake backup data').decode()