# Misapy: a few dev tools for the Misakey stack

Requirements: [`requests` library](https://requests.readthedocs.io/en/master/)

Misapy is a Python *package* with several *modules* in it.
A “Python module” is a directory with a `__main__.py` in it.

To get an access token, you can run `python3 -m misapy.get_access_token`.
You must be located next to the `misapy/` directory for this to work,
meaning that you must see `misapy` when you do `ls`.

There is another module that runs automated tests, `misapy.test`.

HTTP requests are logged to a file in `/tmp`
which path is given during at the beginning of the execution.
This should be quite useful for debugging.