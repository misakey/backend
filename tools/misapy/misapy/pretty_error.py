import traceback
import sys

import requests

from .http import UnexpectedResponseStatus, pretty_string_of_response
from .check_response import BadResponse

# About context managers, see https://www.python.org/dev/peps/pep-0343/
class prettyErrorContext:
    '''A context manager that intercepts exceptions and attemps to pretty-print them'''
    def __init__(self):
        pass

    def __enter__(self):
        pass

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_val:
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
                else:
                    print('Error: Bad Reponse')

                print()
                print(pretty_string_of_response(exc_val.response))

                sys.exit(1)
            # In a context manager,
            # not returning anything indicates that we want the exception to propagate
            # (thus it will be handled by Python as if this context manager did not exist)
