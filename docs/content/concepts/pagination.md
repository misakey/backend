---
title: Pagination
---

## Parameters

Pagination is currently handled using an `limit` and an `offset` integers.

In case of HTTP calls, these command information is received using query parameters.

On endpoint offering pagination, it is usual to have default value for both of these parameters:
- `offset` default value is always 0.
- `limit` default value depends of the endpoint.

The `limit` parameter sets the maximum number of entities contained by the returned list. It is the size of the page.

The `offset` parameter sets the cursor in the total list of stored entities. It is kind of the page number.

### Example:

Considering the list of numbers: **[10, 32, 554, 6, 0.1]**.

- `limit = 0 & offset = 0` --> `[10, 32, 554, 6, 0.1]`
- `limit = 2 & offset = 0` --> `[10, 32]`
- `limit = 2 & offset = 1` --> `[32, 554]`
- `limit = 2 & offset = 3` --> `[6, 0.1]`
- `limit = 2 & offset = 3` --> `[6, 0.1]`
- `limit = 2 & offset = 4` --> `[0.1]` _we have only one record to return_
- `limit = 2 & offset = 5` --> `[]` _no record to return_
- `limit = 10 & offset = 1` --> `[32, 554, 6, 0.1]`

## How to know the total count ?

The total count of existing entities is usually available as on the same route using the HEAD verb instead of the GET.
A Response Header `X-Total-Count` is then returned as an integer.

It might exist some listing endpoints not having its HEAD equivalent implemented, please contact us if you wish it.

## When to stop calling the endpoint if I want to retrieve all the existing entities ?

The consumer should stop calling the list when the number of returned entities is lower than the sent limit.