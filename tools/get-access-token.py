#!/usr/bin/env python3

import argparse
import sys

from misapy.get_access_token import get_credentials
from misapy.test_context import testContext

parser = argparse.ArgumentParser()
parser.add_argument('--email')
args = parser.parse_args()

# TODO extract error mgmt out of "testContext" because it's not really a test here
with testContext():
    creds = get_credentials(args.email)
    
print('email:', creds.email)
print('identity_id:', creds.identity_id)
print('consent has been done' if creds.consent_done else 'no consent required')
print('access token:', creds.access_token)
print('id token:', creds.id_token)
