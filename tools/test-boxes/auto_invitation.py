#!/usr/bin/env python3
from misapy import http, URL_PREFIX
from misapy.box_helpers import create_add_invitation_link_event
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.box_members import join_box
from misapy.test_context import testContext

with testContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)
    # s3 will **not** have an identity pubkey
    s3 = get_authenticated_session(acr_values=2)

    r = s1.post(
        f'{URL_PREFIX}/boxes',
        json={
            'public_key': 'ShouldBeUnpaddedUrlSafeBase64',
            'title': 'Test Box',
        },
    )
    box_id = r.json()['id']

    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [create_add_invitation_link_event()],
        },
    )

    s2.set_identity_pubkey('s2pubkey')

with testContext('"Bad Request" if "auto_invite" but no crypto actions data'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email,
                        'auto_invite': True,
                    },
                    'for_server_no_store': None,
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

with testContext('"Bad Request" if crypto actions data but no "auto_invite"'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email,
                    },
                    'for_server_no_store': {
                        's2pubkey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

with testContext('"Bad Request" if too many keys'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email,
                        'auto_invite': True,
                    },
                    'for_server_no_store': {
                        's2pubkey': 'FakeEncryptedCryptoAction',
                        'badKey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

with testContext('"Bad Request" if missing keys'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email,
                        'auto_invite': True,
                    },
                    'for_server_no_store': {
                        # Note that there is the proper number of keys (1)
                        'badKey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

with testContext('auto invitation'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s2.email,
                        'auto_invite': True,
                    },
                    'for_server_no_store': {
                        's2pubkey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_CREATED,
    )

    r = s2.get(f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 1)
        ]
    )
    cryptoaction = r.json()[0]

    r = s2.get(f'{URL_PREFIX}/identities/{s2.identity_id}/notifications')
    check_response(
        r,
        [
            lambda r: assert_fn(len(r.json()) == 2), # 2 notifs: account creation and auto invitation
            lambda r: assert_fn(r.json()[0]['type'] == 'box.auto_invite'),
            lambda r: assert_fn(r.json()[0]['details']['box_id'] == box_id),
            lambda r:assert_fn(r.json()[0]['details']['cryptoaction_id'] == cryptoaction['id'])
        ]
    )

with testContext('"Conflict" if an identity does not have a public key'):
    s1.post(
        f'{URL_PREFIX}/boxes/{box_id}/batch-events',
        json={
            'batch_type': 'accesses',
            'events' : [
                {
                    'type': 'access.add',
                    'content': {
                        'restriction_type': 'identifier',
                        'value': s3.email,
                        'auto_invite': True,
                    },
                    'for_server_no_store': {
                        'whateverKey': 'becauseS3doesnothaveone',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_CONFLICT,
    )