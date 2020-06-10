import argparse
import sys

parser = argparse.ArgumentParser(prog="misapy")
subparsers = parser.add_subparsers(dest='cmd', required=True)

parser_get_access_token = subparsers.add_parser('get-access-token')
parser_get_access_token.add_argument('--email')

args = parser.parse_args()

if args.cmd == 'get-access-token':
    from .get_access_token import get_credentials
    creds = get_credentials(args.email)
    print('email:', creds.email)
    print('access token:', creds.access_token)
    print('id token:', creds.id_token)
