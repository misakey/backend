+++
categories = ["Concepts"]
date = "2020-09-11"
description = "Formatting of errors"
tags = ["concepts", "formatting", "errors"]
title = "Formatting of errors"
+++

## 1. Error conventions

Backend errors follow these two conventions:
- Always return the HTTP error code that fits the situation.
- Return a JSON object of this shape:
```json
{
     "code": "{Code}",
     "desc": "free format description",
     "origin": "{Origin}",
     "details": {
       "{DetailKey}": "{DetailValue}",
       "{DetailKey}": "{DetailValue}",
     },
}
```

All "{things}" are described in following sections...

## 2. Code

The code is a string used internally and by consumer for better error identifications & reactions.

It is defined by an exhaustive list of values and it corresponds for most of them to HTTP statuses found in the [Mozilla documentation][].

We use this link as general specifications for errors and not just for http request errors.

**Classic codes:** _codes corresponding to http status codes_
* `bad_request`: 400, bad request - wrong request format.
* `unauthorized`: 401, a required valid token is missing/malformed/expired.
* `forbidden`: 403, accesses checks failed - action impossible.
* `not_found`: 404, the resource/route has been not found.
* `method_not_allowed`: 405, the HTTP verb is not supported for the requested route.
* `conflict`: 409, the action cannot be perform considering the current state of the server.
* `unprocessable_entity`: 422, the received entity is unprocessable.
* `internal`: 500, something internal to the service has failed.
* `...`

**Redirect codes**: _code encountered in query parameter `error_code` on redirections_
* `invalid_flow`: the authorization server has raised en error, please check the description to know more about it.
* `login_required`: while an auth flow was inited with `prompt=none` parameter but could not be respected because of authentication is required.
* `consent_required`: while an auth flow was inited with `prompt=none` parameter but could not be respected because of consent is required.
* `missing_parameter`: a required parameter is missing from the request.
* `internal`
* `...`

**Special codes:** _special codes that should not be encountered externally_
:warning: Thanks to contact the backend team if you receive one these codes.
* `unknown_code`: 500, something internal to the service has failed.
* `no_code`: xxx, no specific code defined.

## 3. Origin
Origin is an information about where the error does come from.

**Possible origins:**
* `body`: body parameter.
* `query`: query parameters.
* `path`: path parameters.
* `headers`: headers.
* `internal`: internal logic.
* `...`

**Special origins:** _special origins that should not be encountered externally_
:warning: Thanks to contact the backend team if you receive one these origins.
* `not_defined`: the error has no origin defined yet

## 4. Details

An object containing a dynamical number of detail objects.

Each detail object is built with a DetailKey and a DetailValue:

-> A DetailKey is an dynamical string representing fields name, query parameters name, resource ids...

-> A DetailValue describes an error code as a string for clearer consumer error identification & reactions. It represents codes related to a detail key.

**Possible detail values:**

* `conflict`: unique, conflict with a state machine logic...
* `malformed`: email format,  ip address format...
* `invalid`: minumum/maximum value/lenght...
* `required`: missing in request...
* `expired`: expired duration...
* `forbidden`: forbidden to update...
* `internal`: internal error occured
* `locked`: cannot be updated
* `not_found`: correspondance has not been found
* `not_supported`: not handled by the running implementation
* `timed_out`: something... timed out
* `unauthorized`: authorization is missing

**Special detail values:**
:warning: Thanks to contact the backend team if you receive one these detail values.
* `unknown`: unknown detail code
* `no_code`: no specific code

:warning: In rare cases, the detail is used to give more information about an expected, required value or an resource id to allow a deeper error handling on consumer side.
In this case, an normal formatted detail about the field goes along with this information to give more context and to still be processed in a generic way if wished.

_Non-exhaustive list of examples:_

_On user backup update:_
```json
    {
        "version": "conflict",
        "expected_version": "1"
    }
```

_On any authenticated routes:_
```json
    {
        "acr": "forbidden",
        "required_acr": "2"
    }
```

[Mozilla documentation]: https://developer.mozilla.org/fr/docs/Web/HTTP/Status
