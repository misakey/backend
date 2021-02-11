#!/usr/bin/env python3

import subprocess
import sys

CMD_WITH_ID_FMT = (
    "UPDATE identity SET pubkey = null, non_identified_pubkey = null "
    "WHERE id = '{identity_id}';"
)

CMD_NO_ID = "UPDATE identity SET pubkey = null, non_identified_pubkey = null;"

def main(identity_id):
    cmd = 'docker exec test-and-run_api_db_1  psql -d sso -U misakey -h localhost -c'.split()
    if identity_id == None:
        cmd +=  [CMD_NO_ID]
    else:
        cmd += [CMD_WITH_ID_FMT.format(identity_id=identity_id)]
        
    proc = subprocess.run(cmd, capture_output=True)
    proc.check_returncode()
    output = proc.stdout.decode()
    return output

if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("usage: remove-identity-pubkeys.py (IDENTITY_ID | --all)")

    arg = sys.argv[1]
    identity_id = None if arg == '--all' else arg

    main(identity_id)