+++
categories = ["Endpoints"]
date = "2020-09-28"
description = "Users endpoints"
tags = ["box", "users", "api", "endpoints"]
title = "Box - Users"
+++

# 1. Realtime endpoints

## 1.1 Getting notifications

This websocket (`wss://`) endpoint open a socket.
Notifications will be shipped through this websocket.

[More info](/concepts/realtime) on the events format.

### 1.1.1 request

```bash
    GET wss://api.misakey.com/box-users/74ee16b5-89be-44f7-bcdd-117f496a90a7/ws?access_token=
```

_Query Parameters:_

- `access_token`: For websockets, the access token is (for now) shipped through query parameters.

### 1.1.2 response

The Websocket Protocol handshake is interpreted by HTTP servers as an Upgrade request.
The responses are similar to HTTP classic responses.

_Code_:
```bash
HTTP 200 OK
```

The websocket is then open.
