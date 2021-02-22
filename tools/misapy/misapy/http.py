'''A thin wrapper around the "requests" library.
It supposed to be used exactly as you would use "requests",
but it logs the requests to a file in "/tmp",
and you can provide a "expected_status_code" during calls'''
import datetime
import json
import os
import sys
import os
from urllib.parse import urlparse

import requests
import urllib3; urllib3.disable_warnings()

STATUS_OK = 200
STATUS_CREATED = 201
STATUS_NO_CONTENT = 204
STATUS_BAD_REQUEST = 400
STATUS_FORBIDDEN = 403
STATUS_NOT_FOUND = 404
STATUS_CONFLICT = 409

LATEST_LOGFILE_LINK = '/tmp/misapy-log-latest'
now = datetime.datetime.now()
suffix = now.strftime('%Y-%m-%dT%H-%M-%S')
logfile = f'/tmp/misapy-log-{suffix}'
with open(logfile, 'w') as f:
    f.write(f'''Logfile started on {now.strftime('%Y-%m-%d %H:%M:%S')}\n''')
    f.write('='*40 + '\n'*3)
try:
    os.remove(LATEST_LOGFILE_LINK)
except FileNotFoundError:
    pass
os.symlink(logfile, LATEST_LOGFILE_LINK)
print('log file:', logfile, f'(or {LATEST_LOGFILE_LINK})')

class UnexpectedResponseStatus(Exception):
    def __init__(self, expected_status_code, response):
        self.expected_status_code = expected_status_code
        # Mimicking errors in the requests lib
        self.response = response

        super().__init__(
            f'expected status {expected_status_code}, '
            f'got {response.status_code}'
        )

def call_request_fn_decorated(fn, *args, expected_status_code=None, raise_for_status=True, csrf_token=None, **kwargs):
    '''`raise_for_status` only has effect when there is no `expected_status_code`'''

    if csrf_token:
        kwargs['headers'] = {
            **kwargs.get('headers', {}),
            'X-CSRF-Token': csrf_token,
        }

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
head = lambda *args, **kwargs: call_request_fn_decorated(requests.head, *args, **kwargs)
get = lambda *args, **kwargs: call_request_fn_decorated(requests.get, *args, **kwargs)
patch = lambda *args, **kwargs: call_request_fn_decorated(requests.patch, *args, **kwargs)
put = lambda *args, **kwargs: call_request_fn_decorated(requests.put, *args, **kwargs)
delete = lambda *args, **kwargs: call_request_fn_decorated(requests.delete, *args, **kwargs)

class Session(requests.Session):
    '''requests.Session but with decorated HTTP methods (get, post, etc ...)'''
    def __init__(self):
        super().__init__()
        self.csrf_token = None

    def update_csrf(self, response: requests.Response):
        for redir in response.history:
            if '_csrf' in redir.cookies:
                self.csrf_token = redir.cookies['_csrf']

        if '_csrf' in response.cookies:
            self.csrf_token = response.cookies['_csrf']
        return response
        
    def post(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().post, *args, **kwargs)
        )

    def head(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().head, *args, **kwargs)
        )

    def get(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().get, *args, **kwargs)
        )

    def patch(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().patch, *args, **kwargs)
        )

    def put(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().put, *args, **kwargs)
        )

    def delete(self, *args, **kwargs):
        kwargs['csrf_token'] = self.csrf_token
        return self.update_csrf(
            call_request_fn_decorated(super().delete, *args, **kwargs)
        )


def pretty_string_of_response(response: requests.Response):
    try:
        response_body = json.dumps(response.json(), indent=4)
    except ValueError:
        # response body is not JSON
        response_body = (
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
        try:
            req_payload = request.body[:20].decode() + '...'
        except AttributeError:
            # request body is not JSON
            req_payload = request.body[:20] + '...' if len(request.body) > 20 else ''
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
    parts.append(response_body)

    return '\n'.join(parts)
