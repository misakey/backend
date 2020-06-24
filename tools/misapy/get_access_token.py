import subprocess
import json
import os
from base64 import b64encode
from binascii import hexlify
from collections import namedtuple
from urllib.parse import urlparse
from urllib.parse import parse_qs as parse_query_string

from . import http


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


def login_flow(s, login_challenge, email):
    r = s.put(
        'https://api.misakey.com.local/identities/authable',
        json={
            'login_challenge': login_challenge,
            'identifier': {
                'value': email,
            }
        }
    )
    identity_id = r.json()['authn_step']['identity_id']
    preferred_method = r.json()['authn_step']['method_name']

    # create the emailed code authn step
    # if the preferred method is not emailed code by default
    if preferred_method != "emailed_code":
        s.post(
            'https://api.misakey.com.local/authn-steps',
            json={
                'login_challenge': login_challenge,
                'authn_step': {
                    'identity_id': identity_id,
                    'method_name': 'emailed_code',
                }
            }
        )

    # retrieve the emailed code then authenticate the user
    emailed_code = get_emailed_code(identity_id)
    r = s.post(
        'https://api.misakey.com.local/auth/login/authn-step',
        json={
            'login_challenge': login_challenge,
            'authn_step': {
                'identity_id': identity_id,
                'method_name': 'emailed_code',
                'metadata': {
                    'code': emailed_code,
                }
            }
        }
    )

    manual_redirection = r.json()['redirect_to']
    return identity_id, manual_redirection


def consent_flow(s, consent_challenge, identity_id):
    r = s.post(
        'https://api.misakey.com.local/auth/consent',
        json={
            'consent_challenge': consent_challenge,
            'identity_id': identity_id,
            'consented_scopes': [
                'tos',
                'privacy_policy'
            ]
        }
    )
    manual_redirection = r.json()['redirect_to']
    return manual_redirection


def get_credentials(email=None):
    if not email:
        email = hexlify(os.urandom(3)).decode() + '-test@misakey.com'

    s = http.Session()
    s.verify = False

    # We expect to get a HTTP 502 Bad Gateway in case the frontend is not up,
    # but we don't care because all we need is the login challenge in the redirection URL
    r = s.get(
        'https://auth.misakey.com.local/_/oauth2/auth',
        params={
            'client_id': 'c001d00d-5ecc-beef-ca4e-b00b1e54a111',
            'redirect_uri': 'https://api.misakey.com.local/auth/callback',
            'response_type': 'code',
            'scope': 'openid tos privacy_policy',
            'state': 'shouldBeRandom',
        },
        raise_for_status=False,
    )

    login_url_query = urlparse(r.request.url).query
    login_challenge = parse_query_string(login_url_query)[
        'login_challenge'][0]
    identity_id, manual_redirection = login_flow(s, login_challenge, email)
    r = s.get(manual_redirection, raise_for_status=False)

    # this redirection leads either to consent flow or the final access token
    r = s.get(r.request.url, raise_for_status=False)

    # detect if the consent flow is required
    consent_done = False
    if r.request.url.startswith('https://app.misakey.com.local/auth/consent') == True:
        consent_done = True
        consent_url_query = urlparse(r.request.url).query
        consent_challenge = parse_query_string(consent_url_query)[
            'consent_challenge'][0]
        manual_redirection = consent_flow(s, consent_challenge, identity_id)
        r = s.get(manual_redirection, raise_for_status=False)

    tokens = parse_query_string(urlparse(r.url).fragment)
    access_token = tokens['access_token'][0]
    id_token = tokens['id_token'][0]

    return namedtuple(
        'OAuth2Creds',
        ['email', 'access_token', 'identity_id', 'id_token', 'consent_done'],
    )(email, access_token, identity_id, id_token, consent_done)


def get_authenticated_session(email=None):
    creds = get_credentials(email)
    session = http.Session()
    session.headers.update({'Authorization': f'Bearer {creds.access_token}'})
    session.email = creds.email
    return session
