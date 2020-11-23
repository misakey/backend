# Dev Tools for the Misakey Stack

`get-access-token.py` will perform an auth flow against the stack
and print the obtained access token
along with many other informations.

## Tests

Directory `tests/` contains Python scripts testing various behaviors of our backend.
They are “integration tests” rather than “unit test”
because they require the entire stack to be up
(they don't mock the DB or other sort of ressources)
and they can take quite a lot of time to execute.

HTTP requests are logged to a file in `/tmp`
which path is given during at the beginning of the execution.

### Setup

These Python scripts make use of the `misapy` Python package located in `misapy/misapy`,
so you will have to install this package if you want to run the tests:

```
# pip will need to be pointed at the directory that contains the "setup.py" file
pip install -e ./misapy
```

Note the `-e` function that tells Pip to create a symlink to the package
instead of copying its code to the installation location.
This way you won't have to to `pip install -e ./misapy` every time Misapy is updated.

### Usage

All Python scripts under `tests/` should be executable
and have the proper shebang (`#!/usr/bin/env python3`)
so they should be executable directly from the command line
and shell auto-completion should be able to help you chose which one you want to run:

```
cedricvr@ermit:~/backend$ tools/tests/boxes/
accesses.py                auto-invitation.py         change-invitation-link.py  __pycache__/
all.py                     basics.py                  messages.py
cedricvr@ermit:~/backend$ tools/tests/boxes/auto-invitation.py
log file: /tmp/misapy-log-2020-11-12T18-11-36 (or /tmp/misapy-log-latest)
Tok - 61ee87eb-b354-4be8-8cea-7f48c9918de0: dIXCdnLebL4cifFreVaFkiDzGVrku4Qk6dqNZDxKNfg.nzLZ65nWNT8JTXIO2Bg3Wh9DFMNbLVB9ESN5TC4xF-g
Tok - 8e068767-ecfa-453d-a7e7-f0708dcef70a: l6q1vTnMGUdQZbt-SQVqh5rIoHMvFDpeMbLxmsIjEXE.ZzHExbHrgWOVwhPPbul5PGyaXDK_Km4U_kOVie3Anv8
Tok - 7523f88d-f112-4620-97c9-4b990221a350: ckO2F8TSdJsUtIe6PybvAK6TKvSfUoed6rbJG8X556o.Wj-M9ye1xk-YJAc3D7aNUhIjJy6VS_QEnfmUDbkw6yU
- "Bad Request" if "auto_invite" but no crypto actions data
- "Bad Request" if crypto actions data but no "auto_invite"
- "Bad Request" if too many keys
- "Bad Request" if missing keys
- auto invitation
- "Conflict" if an identity does not have a public key
```

In each directory (except the ones creates by Python like `__pycache__`)
there should be a file `all.py`:
it is a script that will execute all test scripts in this directory
**as well as the ones under it**.

**Known issue with `all.py` scripts:** ([issue #219](https://gitlab.misakey.dev/misakey/backend/-/issues/219))the stack trace is not very useful in determining which instruction failed. If Python raises an exception and you want to see where it comes from, locate which test script it comes from (`all.py` turns test script names into Markdown headers) and run the script directly.

### Contributing

See `tests/README.md`