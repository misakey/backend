'''A thin wrapper around the "requests" library.
It supposed to be used exactly as you would use "requests",
but it logs the requests to a file in "/tmp",
and you can provide a "expected_status_code" during calls'''
import datetime
import json
import sys
from urllib.parse import urlparse

import requests
import urllib3; urllib3.disable_warnings()

now = datetime.datetime.now()
suffix = now.isoformat(timespec='seconds')
logfile = f'/tmp/misapy-log-{suffix}'
print('log file:', logfile)

class UnexpectedResponseStatus(Exception):
    def __init__(self, expected_status_code, response):
        self.expected_status_code = expected_status_code
        # Mimicking errors in the requests lib
        self.response = response

        super().__init__(
            f'expected status {expected_status_code}, '
            f'got {response.status_code}'
        )

def call_request_fn_decorated(fn, *args, expected_status_code=None, raise_for_status=True, **kwargs):
    '''`raise_for_status` only has effect when there is no `expected_status_code`'''
    try:
        response = fn(*args, verify=False, **kwargs)
    except requests.exceptions.ConnectionError as error:
        host = urlparse(error.request.url).netloc
        sys.exit(f'Connection error: is "{host}" up?')

    with open(logfile, 'a') as f:
        f.write(pretty_string_of_response(response))
        f.write('\n\n' + '-'*30 + '\n\n')

    if expected_status_code:
        if response.status_code == expected_status_code:
            return response

        if (
            expected_status_code < 400
            and response.status_code >= 400
        ):
            # should raise
            response.raise_for_status()
        else:
            raise UnexpectedResponseStatus(expected_status_code, response)
    elif raise_for_status:
        response.raise_for_status()

    return response

post = lambda *args, **kwargs: call_request_fn_decorated(requests.post, *args, **kwargs)
patch = lambda *args, **kwargs: call_request_fn_decorated(requests.patch, *args, **kwargs)
put = lambda *args, **kwargs: call_request_fn_decorated(requests.put, *args, **kwargs)
get = lambda *args, **kwargs: call_request_fn_decorated(requests.get, *args, **kwargs)

class Session(requests.Session):
    '''requests.Session but with decorated HTTP methods (get, post, etc ...)'''
    def get(self, *args, **kwargs):
        return call_request_fn_decorated(super().get, *args, **kwargs)
    def post(self, *args, **kwargs):
        return call_request_fn_decorated(super().post, *args, **kwargs)
    def put(self, *args, **kwargs):
        return call_request_fn_decorated(super().put, *args, **kwargs)

def pretty_string_of_response(response):
    try:
        body = json.dumps(response.json(), indent=4)
    except ValueError:
        # response body is not JSON
        body = (
            response.text[:20]
            + '...' if len(response.text) > 20 else ''
        )

    if response.history:
        request = response.history[0].request
    else:
        request = response.request

    if request.headers.get('Content-Type') == 'application/json':
        req_payload = json.dumps(
            json.loads(request.body),
            indent=4
        )
    elif request.body:
        req_payload = request.body[:20].decode() + '...'
    else:
        req_payload = None

    parts = list() 
    parts.append(f'{request.method} {request.url}')
    for (name, value) in request.headers.items():
        parts.append(f'{name}: {value}')
    if req_payload:
        parts.append('Request Body:')
        parts.append(str(req_payload))
    else:
        parts.append('(No Request Body)')
    

    for redir in response.history:
        parts.append(f'\nHTTP {redir.status_code} {redir.reason}')
        for (name, value) in redir.headers.items():
            parts.append(f'{name}: {value}')

    parts.append(f'\nHTTP {response.status_code} {response.reason}')
    for (name, value) in response.headers.items():
        parts.append(f'{name}: {value}')    
    parts.append(body)

    return '\n'.join(parts)
