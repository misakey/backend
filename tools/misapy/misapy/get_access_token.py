import subprocess
import json
import os
from base64 import b64encode
from binascii import hexlify
from collections import namedtuple
from urllib.parse import urlparse
from urllib.parse import parse_qs as parse_query_string

import requests
from . import URL_PREFIX, AUTH_URL_PREFIX, http
from .sessions import Session
from .check_response import check_response, assert_fn
from .password_hashing import hash_password
from .container_access import get_emailed_code
from .utils.base64 import urlsafe_b64encode
from .secret_storage import random_secret_storage_reset_data, random_secret_storage_full_data

def new_password_hash(password):
    salt_base_64 = b64encode(os.urandom(8)).decode()
    return hash_password('password', salt_base_64)

def login_flow(s, login_challenge, email, reset_password=False, use_secret_backup=False, get_secret_storage=False):
    r = s.put(
        f'{URL_PREFIX}/auth/identities',
        json={
            'login_challenge': login_challenge,
            'identifier_value': email,
            'password_reset': reset_password
        }
    )
    identity_id = r.json()['authn_step']['identity_id']
    preferred_method = r.json()['authn_step']['method_name']

    # create the emailed code authn step
    # if the preferred method is not emailed code by default
    if preferred_method != "emailed_code":
        s.post(
            f'{URL_PREFIX}/authn-steps',
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
    r = s.post(
        f'{URL_PREFIX}/auth/login/authn-step',
        json=confirmation_payload
    )

    # In case of a password reset, we need another step here
    if reset_password:
        reset_payload = {
        'login_challenge': login_challenge,
        'authn_step': {
            'identity_id': identity_id,
            'method_name': 'reset_password',
            'metadata': {
                'prehashed_password': new_password_hash('password'),
                'secret_storage': random_secret_storage_reset_data(),
                }
            }
        }
        r = s.post(
            f'{URL_PREFIX}/auth/login/authn-step',
            json=reset_payload
        )

    if r.json()['next'] == 'redirect':
        manual_redirection = r.json()['redirect_to']
        return identity_id, manual_redirection

    # Account creation is required

    assert r.json()['next'] == 'authn_step'
    assert r.json()['authn_step']['method_name'] == 'account_creation'

    if use_secret_backup:
        metadata_secrets = {
            'backup_data': b64encode(b'fake backup data').decode(),
        }
    else:
        metadata_secrets = {
            'secret_storage': random_secret_storage_reset_data(),
        }


    r = s.post(
        f'{URL_PREFIX}/auth/login/authn-step',
        cookies={"authnaccesstoken": s.cookies['authnaccesstoken'], "authntokentype": "bearer"},
        json={
            'login_challenge': login_challenge,
            'authn_step': {
                'identity_id': identity_id,
                'method_name': 'account_creation',
                'metadata': {
                    'prehashed_password': new_password_hash('password'),
                    **metadata_secrets,
                }
            }
        }
    )
    
    manual_redirection = r.json()['redirect_to']

    if get_secret_storage:
        r = s.get(
            f'{URL_PREFIX}/auth/secret-storage',
            params={
                'login_challenge': login_challenge,
                'identity_id': identity_id,
            },
        cookies={"authnaccesstoken": s.cookies['authnaccesstoken'], "authntokentype": "bearer"},        )
        check_response(
            r,
            [
                lambda r: assert_fn(set(r.json().keys()) == {'secrets', 'account_id'}),
            ]
        )

    return identity_id, manual_redirection

def consent_flow(s, consent_challenge, identity_id):
    r = s.post(
        f'{URL_PREFIX}/auth/consent',
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


def get_user_credentials(email=None, require_account=False, acr_values=None, reset_password=False, use_secret_backup=False, get_secret_storage=False):
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

    s = Session()
    s.verify = False

    # We expect to get a HTTP 502 Bad Gateway in case the frontend is not up,
    # but we don't care because all we need is the login challenge in the redirection URL
    r = s.get(
        f'{AUTH_URL_PREFIX}/_/oauth2/auth',
        params={
            'client_id': 'cc411b8f-28bf-4d4e-abd9-99226b41da27',
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
            lambda r: assert_fn('error=' not in r.history[0].headers['location'])
        ]
    )

    login_url_query = urlparse(r.request.url).query
    login_challenge = parse_query_string(login_url_query)['login_challenge'][0]
    identity_id, manual_redirection = login_flow(s, login_challenge, email, reset_password, use_secret_backup, get_secret_storage)
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
        f'{URL_PREFIX}/identities/{identity_id}',
        cookies={"accesstoken": access_token, "tokentype": "bearer"},
    )
    account_id = r.json()['account_id']
    display_name = r.json()['display_name']

    s.identity_id = identity_id
    s.email = email
    s.account_id = account_id
    s.display_name = display_name


    return namedtuple(
        'OAuth2Creds',
        ['email', 'access_token', 'identity_id',
            'id_token', 'consent_done', 'account_id', 'display_name',
            'session'],
    )(email, access_token, identity_id, id_token, consent_done, account_id, display_name, s)


def get_authenticated_session(email=None, require_account=False, acr_values=None, reset_password=False, use_secret_backup=False, get_secret_storage=False):
    creds = get_user_credentials(email, require_account, acr_values, reset_password, use_secret_backup, get_secret_storage)
    print(f'Tok - {creds.identity_id}: {creds.access_token}')
    return creds.session


def perform_org_auth_flow(org_id, org_secret):
    s = Session()
    r = s.post(
        f'{AUTH_URL_PREFIX}/_/oauth2/token',
        data={
            'grant_type': 'client_credentials',
            'scope': '',
            'client_id': org_id,
            'client_secret': org_secret,
        },
    )

    s.org_access_token=r.json()['access_token']
    s.org_id = org_id
    return s


def get_org_session(user_session):
    print('- user creates an organization')
    name = hexlify(os.urandom(3)).decode() + '-org'
    r = user_session.post(
        f'{URL_PREFIX}/organizations',
        json={'name': name},
        expected_status_code=http.STATUS_CREATED
    )
    check_response(r,
        [
            lambda r: assert_fn(r.json()['name'] == name),
            lambda r: assert_fn(r.json()['creator_id'] == user_session.identity_id),
            lambda r: assert_fn(r.json()['current_identity_role'] == 'admin'),
        ]
    )
    org_id = r.json()['id']

    print(f'- user generates secret for the org {org_id}')
    r = user_session.put(
        f'{URL_PREFIX}/organizations/{org_id}/secret',
        expected_status_code=http.STATUS_OK
    )

    print(f'- generate the auth session for the org {org_id}')
    s = perform_org_auth_flow(org_id, r.json()['secret'])
    print(f'Org Tok - {s.org_id}: {s.org_access_token}')
    s.org_name = name
    return s
