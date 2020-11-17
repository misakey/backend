# How to Contribute to Tests Scripts

Files in this directory are Python scripts.

They must be marked as executable (`chmod u+x`)
and start with the shebang `#!/usr/bin/env python3`
so that they can be invoked from the command line
without having to prefix the command with `python3`.

Each test script is supposed to be runnable independently of the others.
This means that each script is supposed to create its own test user accounts,
for instance.

For file names, please prefer `kebab-case.py`:
this is the usual practice for executable names,
it is easier to type on QWERTY keyboards,
and it has the nice property of not playing well with Python's `import` keyword,
which is good because these scripts are supposed to be *executed*, not *imported*.

Scripts `all.py` will execute *“any Python script underneath them”* so you don't have to “add” your new script to any list for it to be executed by `all.py`.

## Making HTTP Requests

Use the methods from `misapy.http`,
or a `Session` object from either `misapy.http`
or as returned by `get_authenticated_session` (provided by `misapy.get_access_token`).

They are thin wrappers around the `requests` library
([documentation](https://requests.readthedocs.io/en/master/))
that mainly add logging as well as the additional keyword argument `expected_response_status`.

## Error Pretty-printing

The `misapy` package (located in `tools/misapy/misapy`)
provides a Python context manager `prettyErrorContext`
which will print information about the HTTP request and response that caused an error.

Simply executing your test script in this context should suffice:

```python
from misapy.pretty_error import prettyErrorContext

with prettyErrorContext():
	# your test script goes here
```

## Making Assertions on HTTP Responses

You can simply use the `assert` keyword if you want,
but there is a `check_response` provided by `misapy.check_response` that produces nicer output
because it will raise an error that can be pretty-printed by `prettyErrorContext`.

Here is how you use it:

```python
check_response(
    r,
    [
        lambda r: assert_fn(r.json()[0]['type'] == 'set_box_key_share')
    ]
)
```

Where `r` is the response and `assert_fn` is provided by `misapy.check_response`
(this is because the `assert` keyword cannot be used in a `lambda`).

Try to have one lambda per line,
so that the entire assertion that fails will be printed by Python in the error stack trace:
in the above example, if the assertion fails you will see a stack trace
including the line `lambda r: assert_fn(r.json()[0]['type'] == 'set_box_key_share')`
(so you can see what assertion failed)
as well as the HTTP responses that did not pass this assertion
(and the request that caused this response).