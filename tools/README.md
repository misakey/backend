# Misapy: a few dev tools for the Misakey stack

Requirements: [`requests` library](https://requests.readthedocs.io/en/master/)

Python scripts in this directory are executables
that you can call from any location on your machine.
Because they have a “shebang”, you don't even have to call the Python interpreter explicitely.
For instance, to get an access token:

    path/to/get-access-token.py [--email whatever.email@you.want]

## Setup

Python scripts located next to `misapy` should work out-of-the-box,
but some are grouped in directories such as `test-boxes/`
and as a result `import misapy` in these scripts will not work out-of-the-box
because Python will not find `misapy` next to “the module we are executing”
(here, the Python file).

The solution is to install Misapy
so that `import misapy` works from anywhere on your system:
do `pip install -e ./` in the present directory
(what matters is that you are in the same directory as `setup.py`).
What's nice with the `-e` option is that it does not copy the current Misapy source code to your Python packages,
it just creates a symlink, so you should not have to re-run `pip install -e ./`
even if modifications are made to Misapy.

If for some reason you don't want to install Misapy
but you still want to run the test scripts that are in directories, there is a trick:
`cd` into the present directory (`tools/`)
and run `python -m test-boxes.change_invitation_link`
(or another file system path written with dots and without  the `.py` extension).
This will tell Python to consider `test-boxes/change_invitation_link.py`
as a submodule of the “module” `test-box`,
so that `import` directives will look for modules next to `test-box/`
instead of looking next to `test-boxes/change_invitation_link.py`.
The drawback is that you *have* to be located into `tools/` when executing that,
and its longer to type than just `test-boxes/change_invitation_link.py`
(with the benefit of shell auto-completion).

## Features

HTTP requests are logged to a file in `/tmp`
which path is given during at the beginning of the execution.
