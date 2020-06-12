# Misapy: a few dev tools for the Misakey stack

Requirements: [`requests` library](https://requests.readthedocs.io/en/master/)

Python scripts in this directory are executables
that you can call from any location on your machine.
Because they have a “shebang”, you don't even have to call the Python interpreter explicitely.
When you execute a Python script, the path used by Python to resolve imports is the directory where the script is located, so `from misapy import [...]` in the scripts will always work no matter where you are located when you call the scripts.

For instance, to get an access token:

    path/to/get-access-token.py [--email whatever.email@you.want]

## Features

HTTP requests are logged to a file in `/tmp`
which path is given during at the beginning of the execution.
