#!/usr/bin/env python3

import argparse
import sys
from pathlib import Path

# so that this script can run even if package "misapy" has not been installed
sys.path.append(str(Path(__file__).parent / 'misapy'))

from misapy.get_access_token import get_credentials
from misapy.pretty_error import prettyErrorContext

parser = argparse.ArgumentParser()
parser.add_argument('--email')
parser.add_argument('--acr')
parser.add_argument('--require-account',
                    dest='require_account', action='store_true')
args = parser.parse_args()

with prettyErrorContext():
    creds = get_credentials(
        args.email, args.require_account, args.acr)

print('email:', creds.email)
print('acr_values:', args.acr)
print('identity_id:', creds.identity_id)
if creds.account_id:
    print('account_id:', creds.account_id)
print('consent has been done' if creds.consent_done else 'no consent required')
print('access token:', creds.access_token)
print('CSRF token:', creds.csrf_token)
print('id token:', creds.id_token)
