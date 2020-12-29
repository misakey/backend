import subprocess
import json
import os
from base64 import b64encode
from binascii import hexlify
from collections import namedtuple
from urllib.parse import urlparse
from urllib.parse import parse_qs as parse_query_string

import requests
from . import http
from .sessions import Session
from .check_response import check_response, assert_fn
from .password_hashing import hash_password
from .container_access import get_emailed_code

def new_password_hash(password):
    salt_base64 = b64encode(os.urandom(8)).decode()
    return hash_password('password', salt_base64)

def login_flow(s, login_challenge, email, reset_password=False):
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
    confirmation_payload = {
        'login_challenge': login_challenge,
        'authn_step': {
            'identity_id': identity_id,
            'method_name': 'emailed_code',
            'metadata': {
                'code': emailed_code,
            }
        }
    }
    if reset_password:
        confirmation_payload['password_reset'] = {
            'prehashed_password': new_password_hash('password'),
            'backup_data': b64encode(b'other fake backup data').decode(),
        }
    r = s.post(
        'https://api.misakey.com.local/auth/login/authn-step',
        json=confirmation_payload
    )

    if r.json()['next'] == 'redirect':
        manual_redirection = r.json()['redirect_to']
        return identity_id, manual_redirection

    # Account creation is required

    assert r.json()['next'] == 'authn_step'
    assert r.json()['authn_step']['method_name'] == 'account_creation'

    # temporary access token
    auth_access_token = r.json()['access_token']

    r = s.post(
        'https://api.misakey.com.local/auth/login/authn-step',
        headers={
            'Authorization': f'Bearer {auth_access_token}'
        },
        json={
            'login_challenge': login_challenge,
            'authn_step': {
                'identity_id': identity_id,
                'method_name': 'account_creation',
                'metadata': {
                    'prehashed_password': new_password_hash('password'),
                    'backup_data': b64encode(b'fake backup data').decode(),
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


def get_credentials(email=None, require_account=False, acr_values=None, reset_password=False):
    '''if no email is passed, a random one will be used.'''

    if require_account:
        if acr_values:
            raise ValueError('cannot use "require_account" and "acr_values"')
        else:
            acr_values = 2

    if reset_password and not email:
        raise ValueError('"reset password" requires an "email" parameter')

    if not email:
        email = hexlify(os.urandom(3)).decode() + '-test@misakey.com'

    s = http.Session()
    s.verify = False

    # We expect to get a HTTP 502 Bad Gateway in case the frontend is not up,
    # but we don't care because all we need is the login challenge in the redirection URL
    r = s.get(
        'https://auth.misakey.com.local/_/oauth2/auth',
        params={
            'client_id': '00000000-0000-0000-0000-000000000000',
            'redirect_uri': 'https://api.misakey.com.local/auth/callback',
            'response_type': 'code',
            'scope': 'openid tos privacy_policy',
            'state': 'shouldBeRandom',
            'acr_values': acr_values,
        },
        raise_for_status=False,
    )

    check_response(
        r,
        [
            lambda r: assert_fn(
                'error=' not in r.history[0].headers['location'])
        ]
    )

    login_url_query = urlparse(r.request.url).query
    login_challenge = parse_query_string(login_url_query)['login_challenge'][0]
    identity_id, manual_redirection = login_flow(s, login_challenge, email, reset_password)
    r = s.get(manual_redirection, raise_for_status=False)

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
    id_token = tokens['id_token'][0]
    # for some reason the cookie does not appear in `r.cookies`
    # but it appears in `s.cookies`
    access_token = s.cookies['accesstoken']

    r = http.get(
        f'https://api.misakey.com.local/identities/{identity_id}',
        cookies={"accesstoken": access_token, "tokentype": "bearer"},
    )
    account_id = r.json()['account_id']
    identifier_id = r.json()['identifier_id']
    display_name = r.json()['display_name']

    s.identity_id = identity_id
    s.email = email
    s.account_id = account_id
    s.identifier_id = identifier_id
    s.display_name = display_name


    return namedtuple(
        'OAuth2Creds',
        ['email', 'access_token', 'identity_id',
            'id_token', 'consent_done', 'account_id', 'identifier_id', 'display_name',
            'session'],
    )(email, access_token, identity_id, id_token, consent_done, account_id, identifier_id, display_name, s)


def get_authenticated_session(email=None, require_account=False, acr_values=None, reset_password=False):
    creds = get_credentials(email, require_account, acr_values, reset_password)
    print(f'Tok - {creds.identity_id}: {creds.access_token}')
    return creds.session
