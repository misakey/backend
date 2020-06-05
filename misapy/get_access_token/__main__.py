import subprocess
import json
import os
from base64 import b64encode
from binascii import hexlify
from urllib.parse import urlparse
from urllib.parse import parse_qs as parse_query_string

from .. import http
import urllib3; urllib3.disable_warnings()

def get_confirmation_code(identity_id):
  proc = subprocess.run(
    (
      'docker exec test-and-run_api_db_1  psql -t -d api -U misakey -h localhost -c'.split()
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
  confirmation_code = json.loads(output)['code']
  return confirmation_code

s = http.Session()
s.verify = False

# We expect to get a HTTP 502 Bad Gateway because the frontend may not be up,
# but we don't care because all we need is the login challenge in the redirection URL
r = s.get(
  'https://auth.misakey.com.local/_/oauth2/auth',
  params={
  'client_id': 'c001d00d-5ecc-beef-ca4e-b00b1e54a111',
  'redirect_uri': 'https://api.misakey.com.local/auth/callback',
  'response_type': 'code',
  'scope': 'openid',
  'state': 'shouldBeRandom',
  },
  expected_status_code=502,
)

last_redirection = r.request.url
login_challenge = parse_query_string(urlparse(last_redirection).query)['login_challenge'][0]

email = hexlify(os.urandom(3)).decode() + '-test@misakey.com'

r = s.put(
  'https://api.misakey.com.local/identities/authable',
  json={
    'login_challenge': login_challenge,
    'identifier': {
      'value': email,
    }
  }
)
r.raise_for_status()
identity_id = r.json()['authn_step']['identity_id']

confirmation_code = get_confirmation_code(identity_id)

r = s.post(
  'https://api.misakey.com.local/auth/login/authn-step',
  json={
    'login_challenge': login_challenge,
    'authn_step': {
      'identity_id': identity_id,
      'method_name': 'emailed_code',
      'metadata': {
        'code': confirmation_code,
      }
    }
  }
)
r.raise_for_status()

manual_redirection = r.json()['redirect_to']

r = s.get(manual_redirection, expected_status_code=502)

tokens = parse_query_string(urlparse(r.url).fragment)
access_token = tokens['access_token'][0]
id_token = tokens['id_token'][0]

print('access token:', access_token)
print('id_token:', id_token)