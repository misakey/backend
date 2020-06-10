'''tools for running various checks on a HTTP response'''

import requests

from .http import UnexpectedResponseStatus

def check_response(response, checks):
  for check in checks:
    try:
      check(response)
    except Exception as error:
      raise BadResponse(response) from error

class BadResponse(Exception):
  def __init__(self, response):
    self.response = response

    request = response.request
    error_msg = f'Bad response for {request.method} {request.url}'

    super().__init__(error_msg)

# because "assert" cannot be used in a lambda
def assert_fn(check):
  assert check

class expectHttpErrorContext:
  def __init__(self, expected_code):
    self.expected_code = expected_code

  def __enter__(self):
    pass

  def __exit__(self, exc_type, exc_val, exc_tb):
    if not exc_val:
      raise Exception('Expected an exception')

    if exc_type in [requests.exceptions.HTTPError, UnexpectedResponseStatus]:
      if exc_val.response.status_code == self.expected_code:
        return True # exception is "swallowed"      