#!/usr/bin/env python3

import argparse
import sys

from misapy.get_access_token import get_credentials

parser = argparse.ArgumentParser()
parser.add_argument('--email')
args = parser.parse_args()

creds = get_credentials(args.email)
print('email:', creds.email)
print('access token:', creds.access_token)
print('id token:', creds.id_token)
