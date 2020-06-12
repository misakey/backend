import traceback
import sys

import requests

from .http import UnexpectedResponseStatus, pretty_string_of_response
from .check_response import BadResponse

# About context managers, see https://www.python.org/dev/peps/pep-0343/
class testContext:
    def __init__(self, test_name=None):
        self.test_name = test_name

    def __enter__(self):
        if self.test_name:
            print('… ' + self.test_name, end='')

    def __exit__(self, exc_type, exc_val, exc_tb):
        if not exc_val:
            if self.test_name:
                print('\r' + '✓ ' + self.test_name)
        else:
            if self.test_name:
                print('\r' + '✗ ' + self.test_name)

            # Pretty printing of some types of errors
            if exc_type in [requests.exceptions.HTTPError, UnexpectedResponseStatus]:
                traceback.print_exception(exc_type, exc_val, exc_tb)

                print()
                print(pretty_string_of_response(exc_val.response))

                sys.exit(1)
            elif exc_type == BadResponse:
                cause = traceback.TracebackException(exc_type, exc_val, exc_tb).__cause__
                if cause:
                    print(''.join(cause.format()))

                print('Caused by response:')
                print(textwrap.indent(
                    pretty_string_of_response(exc_val.response),
                    ' '*2,
                ))

                sys.exit(1)
            # In a context manager,
            # not returning anything indicates that we want the exception to propagate
