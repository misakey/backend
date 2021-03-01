#!/usr/bin/env python3

from misapy import http, URL_PREFIX
from misapy.box_helpers import create_box_with_data_subject_and_datatag
from misapy.org_helpers import create_org, create_datatag
from misapy.box_members import join_box
from misapy.check_response import check_response, assert_fn
from misapy.get_access_token import get_authenticated_session
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
    s1 = get_authenticated_session(acr_values=2)
    s2 = get_authenticated_session(acr_values=2)

    oid = create_org(s1)
    did = create_datatag(s1, oid)
    oid2 = create_org(s1)
    did2 = create_datatag(s1, oid2)

    print("- create box with org")
    box_id = create_box_with_data_subject_and_datatag(s1, oid)
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['owner_org_id'] == oid 
    assert r.json()['access_mode'] == 'limited'

    print("- create box with org and datatag")
    box_id = create_box_with_data_subject_and_datatag(s1, org_id=oid, datatag_id=did)
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['owner_org_id'] == oid
    assert r.json()['datatag_id'] == did
    assert r.json()['access_mode'] == 'limited'


    print("- forbidden if create box with datatag but without org")
    box_id = create_box_with_data_subject_and_datatag(s1, datatag_id=did, expected_status_code=403)

    print("- forbidden if create box with datatag not in org")
    box_id = create_box_with_data_subject_and_datatag(s1, org_id=oid, datatag_id=did2, expected_status_code=403)

    print("- create box with data subject with no account")
    box_id = create_box_with_data_subject_and_datatag(s1, data_subject="awesome_user@example.com")
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['access_mode'] == 'limited'

    r = s1.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value=awesome_user@example.com',
        expected_status_code=200,
    )
    # there shouldn’t be any pubkey since the user has just been created
    assert r.json() == []

    print("- bad request if create box with data subject with account but missing invitation data")
    box_id = create_box_with_data_subject_and_datatag(s1, data_subject=s2.email, expected_status_code=400)

    print("- create box with data subject with account")
    r = s1.get(
        f'{URL_PREFIX}/identities/pubkey?identifier_value={s2.email}',
        expected_status_code=200,
    )
    pubkey = r.json()[0]

    box_id = create_box_with_data_subject_and_datatag(s1, data_subject=s2.email, public_key=pubkey)

    # check crypto action
    r = s2.get(
        f'{URL_PREFIX}/accounts/{s2.account_id}/crypto/actions',
        expected_status_code=200,
    )
    assert r.json()[0]['encrypted'] == 'fakeCryptoAction'

    # try to join
    r = s2.join_box(box_id)

    # check box state
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}',
        expected_status_code=200,
    )
    assert r.json()['title'] != ''
    assert r.json()['access_mode'] == 'limited'
    assert r.json()['subject']['id'] == s2.identity_id

    # check box members
    r = s1.get(
        f'{URL_PREFIX}/boxes/{box_id}/members',
        expected_status_code=200,
    )
    identity_ids = [elt['id'] for elt in r.json()]
    assert s2.identity_id in identity_ids
    assert s1.identity_id in identity_ids

    print("- bad request if create box with data subject with account and wrong pubkey")

    box_id = create_box_with_data_subject_and_datatag(s1, data_subject=s2.email, public_key="wrongPubkey", expected_status_code=400)
