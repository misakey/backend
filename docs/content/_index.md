+++
categories = ["Misakey"]
date = "2020-09-11"
description = "Misakey API Documentation"
tags = ["misakey", "api", "documentation"]
title = "Misakey API Documentation"
+++

Welcome to our API documentation.

**Remark:** most of the API is described by examples
more than by rigorous formal specification.
For instance when you see a UUID somewhere,
you should understand that the value at this location must be an UUID,
not that it must be *this exact UUID* (which wouldn't make sense anyway).

Global rules:
* All time strings are represented following the [RFC 3339](https://tools.ietf.org/html/rfc3339).
* All request and response parameters described in following specifications are **required** or **always returned**,
unless other rules are locally specified.
