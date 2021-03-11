#!/usr/bin/env python3

from misapy import http, URL_PREFIX
from misapy.box_helpers import create_org_box
from misapy.org_helpers import create_org, create_datatag
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session, get_org_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    sorg1 = get_org_session(s1)
    oid = sorg1.org_id
    oid2 = create_org(s1)
    did = create_datatag(s1, oid)
    did2 = create_datatag(s1, oid2)


    print('- bad request if create box without data subject')
    error = create_org_box(sorg1, datatag_id=did, expected_status_code=400)
    check_response(error, [lambda r: assert_fn(error['details']['data_subject'] == 'required')])

    print('- bad request if create box without datatag')
    error = create_org_box(sorg1, data_subject='t@t.t', expected_status_code=400)
    check_response(error, [lambda r: assert_fn(error['details']['datatag_id'] == 'required')])

    print('- forbidden if create box with datatag not in org')
    create_org_box(sorg1, org_id=oid, datatag_id=did2, data_subject='t@t.t', expected_status_code=403)

    print('- create box with data subject with no account')
    box = create_org_box(sorg1, data_subject='awesome_user@example.com', datatag_id=did)
    box_id = box['id']
    r = sorg1.get(
        f'{URL_PREFIX}/organizations/{oid}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['access_mode'] == 'limited'
    assert r.json()['datatag_id'] == did

    print('- there should not be any pubkey since the user has just been created')
    r = s1.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value=awesome_user@example.com',
        expected_status_code=200,
    )
    assert r.json() == []

    print('- bad request if create box with data subject with account but missing invitation data')
    error = create_org_box(sorg1, datatag_id=did, data_subject=s1.email, expected_status_code=400)
    check_response(error, [lambda r: assert_fn(error['details']['invitation_data'] == 'required')])
    
    print('- create box with data subject with account')
    r = sorg1.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={s1.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]
    box = create_org_box(sorg1, datatag_id=did, data_subject=s1.email, public_key=pubkey)
    box_id = box['id']

    print('- crypto action has been created')
    r = s1.get(
        f'{URL_PREFIX}/accounts/{s1.account_id}/crypto/actions',
        expected_status_code=200,
    )
    assert r.json()[0]['encrypted'] == 'fakeCryptoAction'

    print('- the data subject can join the box')
    r = s1.join_box(box_id)

    print('- the data subject can then get the box')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['access_mode'] == 'limited'
    assert r.json()['subject']['id'] == s1.identity_id

    print('- the data subject can see org and themself in members')
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200,
    )
    identity_ids = [elt['id'] for elt in r.json()]
    assert s1.identity_id in identity_ids
    assert oid in identity_ids

    print('- bad request if create box with data subject with account and wrong pubkey')
    create_org_box(sorg1, datatag_id=did, data_subject=s1.email, public_key='wrongPubkey', expected_status_code=400)
