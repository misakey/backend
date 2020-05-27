'''A thin wrapper around the "requests" library
to add a few custom behavior'''
import datetime
import json

import requests

now = datetime.datetime.now()
suffix = now.isoformat(timespec='seconds')
logfile = f'/tmp/api-test-log-{suffix}'
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

def wrap_requests_method(method):
    def wrapped(*args, expected_status_code=None, **kwargs):
        response = getattr(requests, method)(*args, **kwargs)

        with open(logfile, 'a') as f:
            f.write(pretty_string_of_response(response))
            f.write('\n\n')

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
        
        response.raise_for_status()
        return response
    return wrapped

post = wrap_requests_method('post')
patch = wrap_requests_method('patch')
put = wrap_requests_method('put')
get = wrap_requests_method('get')

def pretty_string_of_response(response):
    try:
        body = json.dumps(response.json(), indent=4)
    except ValueError:
        # response body is not JSON
        body = response.text

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
    if req_payload:
        parts.append('Request Body:')
        parts.append(str(req_payload))
    else:
        parts.append('(No Request Body)')
    parts.append('\nResponse')
    parts.append(f'HTTP {response.status_code} {response.reason}')
    parts.append(body)

    return '\n'.join(parts)
