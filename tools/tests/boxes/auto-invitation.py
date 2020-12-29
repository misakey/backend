#!/usr/bin/env python3
from misapy import http, URL_PREFIX
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.box_members import join_box
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
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

    s2.set_identity_pubkey('s2pubkey')
    
    print('- "Bad Request" if "auto_invite" but no crypto actions data')
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
                    'extra': None,
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

    print('- "Bad Request" if crypto actions data but no "auto_invite"')
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
                    'extra': {
                        's2pubkey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

    print('- "Bad Request" if too many keys')
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
                    'extra': {
                        's2pubkey': 'FakeEncryptedCryptoAction',
                        'badKey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

    print('- "Bad Request" if missing keys')
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
                    'extra': {
                        # Note that there is the proper number of keys (1)
                        'badKey': 'FakeEncryptedCryptoAction',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_BAD_REQUEST,
    )

    print('- auto invitation')
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
                    'extra': {
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
    notification = r.json()[0]
    check_response(
        r,
        [
            lambda r: assert_fn(notification['type'] == 'box.auto_invite'),
            lambda r: assert_fn(notification['details']['box_id'] == box_id),
            lambda r:assert_fn(notification['details']['cryptoaction_id'] == cryptoaction['id'])
        ]
    )

    print('- "Conflict" if an identity does not have a public key')
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
                    'extra': {
                        'whateverKey': 'becauseS3doesnothaveone',
                    }
                }
            ]
        },
        expected_status_code=http.STATUS_CONFLICT,
    )

    print('- deletion of cryptoaction')
    s2.delete(
        f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions/{cryptoaction["id"]}',
        expected_status_code=http.STATUS_NO_CONTENT,
    )

    r = s2.get(f'{URL_PREFIX}/identities/{s2.identity_id}/notifications')
    # finding the previous notification
    notification = [
        x for x in r.json()
        if x['id'] == notification['id']
    ][0]
    check_response(
        r,
        [
            # Must have been marked as used
            lambda r: assert_fn(notification['details']['used'] == True),
        ]
    )