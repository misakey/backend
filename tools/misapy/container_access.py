import subprocess
import json

def get_emailed_code(identity_id):
    proc = subprocess.run(
        (
            'docker exec test-and-run_api_db_1  psql -t -d sso -U misakey -h localhost -c'.split()
            + [
                "SELECT metadata "
                "FROM authentication_step "
                f"WHERE identity_id = '{identity_id}' "
                "ORDER BY created_at DESC LIMIT 1;"
            ]
        ),
        capture_output=True,
    )
    proc.check_returncode()
    output = proc.stdout.decode()
    emailed_code = json.loads(output)['code']
    return emailed_code

def list_encrypted_files():
    proc = subprocess.run(
        (
            'docker exec -t test-and-run_api_1 sh -c'.split()
            + ["ls /etc/encrypted-files"]
        ),
        capture_output=True,
    )
    proc.check_returncode()
    output = proc.stdout.decode()
    file_ids = output.split()
    return file_ids