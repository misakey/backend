#!/usr/bin/env python3

import subprocess

def main():
    proc = subprocess.run(
        (
            'docker exec test-and-run_api_db_1  psql -d sso -U misakey -h localhost -c'.split()
            + [
                "DELETE FROM secret_storage_account_root_key;"
                "DELETE FROM secret_storage_asym_key;"
                "DELETE FROM secret_storage_vault_key;"
                "DELETE FROM secret_storage_box_key_share;"
            ]
        ),
        capture_output=True,
    )
    proc.check_returncode()
    output = proc.stdout.decode()
    return output

if __name__ == "__main__":
    main()