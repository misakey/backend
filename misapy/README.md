# Misapy: a few dev tools for the Misakey stack

Requirements: [`requests` library](https://requests.readthedocs.io/en/master/)

To get an access token:

    python3 -m misapy get-access-token [--email some@email.com]

You must be located next to the `misapy/` directory for this to work,
meaning that you must see `misapy` when you do `ls`.

HTTP requests are logged to a file in `/tmp`
which path is given during at the beginning of the execution.
This should be quite useful for debugging.