# MonoHook

Expose a single HTTP webhook endpoint that executes a command:

```
$> monohook --authorization letmein --concurrency 1 --port 5000 -- deploy.sh
```

Trigger execution of the command by POSTing to the webhook:

```
$> curl -v -X POST -H "Authorization: Bearer letmein" http://localhost:5000
> POST / HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
> Authorization: Bearer letmein
>
< HTTP/1.1 202 Accepted
< Date: Wed, 15 May 2019 07:06:30 GMT
< Content-Length: 0
```

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![version](https://img.shields.io/badge/Version-v2.0.0-blue.svg)](https://github.com/AlphaHydrae/monohook/releases/tag/v2.0.0)
[![build status](https://travis-ci.org/AlphaHydrae/monohook.svg?branch=master)](https://travis-ci.org/AlphaHydrae/monohook)
[![license](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.txt)



## Installation

### Docker

Run it with Docker:

```
$> docker run --rm -it -p 5000:5000 alphahydrae/monohook:2.0.0 -a letmein -- echo Hook received
```

Test it:

```
$> curl -v -X POST -H "Authorization: Bearer letmein" http://localhost:5000
```

You should see the command being executed in the container's logs:

```
Executing /bin/echo
Hook received
Successfully executed /bin/echo
```

To use MonoHook in a `Dockerfile`, simply download the binary and make it executable:

```
RUN wget -O /usr/local/bin/monohook \
    https://github.com/AlphaHydrae/monohook/releases/download/v2.0.0/monohook_linux_amd64 && \
    chmod +x /usr/local/bin/monohook
```

### Download binary

* **Linux**

  ```
  wget -O /usr/local/bin/monohook \
    https://github.com/AlphaHydrae/monohook/releases/download/v2.0.0/monohook_linux_amd64 && \
    chmod +x /usr/local/bin/monohook
  ```
* **Linux (arm64)**

  ```
  wget -O /usr/local/bin/monohook \
    https://github.com/AlphaHydrae/monohook/releases/download/v2.0.0/monohook_linux_arm64 && \
    chmod +x /usr/local/bin/monohook
  ```
* **macOS**

  ```
  wget -O /usr/local/bin/monohook \
    https://github.com/AlphaHydrae/monohook/releases/download/v2.0.0/monohook_darwin_amd64 && \
    chmod +x /usr/local/bin/monohook
  ```
* **Windows**

  ```
  wget -O /usr/local/bin/monohook \
    https://github.com/AlphaHydrae/monohook/releases/download/v2.0.0/monohook_windows_amd64 && \
    chmod +x /usr/local/bin/monohook
  ```



## Usage

MonoHook listens on a given TCP port, 5000 by default, and executes a command
when a request is received. The command must be specified after the argument
terminator (`--`):

```
$> monohook -- echo Hook received
```

To trigger execution of the command, simply make a POST request to the hook:

```
$> curl -v -X POST http://localhost:5000
```

If you go back to the terminal where MonoHook is running, you will see that the
command's output is piped to MonoHook's own standard output so you can see it:

```
Executing /bin/echo
Hook received
Successfully executed /bin/echo
```

### Authentication

You can secure the webhook by requiring that an authentication token be sent,
using the `-a, --authorization <TOKEN>` command-line flag or the
`$MONOHOOK_AUTHORIZATION` environment variable:

```
$> monohook -a letmein -- echo Hook received
```

MonoHook will respond to requests without authentication with HTTP 401
Unauthorized, and the command will not be executed:

```
$> curl -v -X POST http://localhost:5000
> POST / HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 401 Unauthorized
< Date: Wed, 15 May 2019 07:22:05 GMT
< Content-Length: 0
```

You may send the authentication token as a bearer token in the `Authorization`
header:

```
$> curl -v -X POST -H "Authorization: Bearer letmein" http://localhost:5000
> POST / HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
> Authorization: Bearer letmein
>
< HTTP/1.1 202 Accepted
< Date: Wed, 15 May 2019 07:24:11 GMT
< Content-Length: 0
```

Or in the `authorization` URL query parameter:

```
$> curl -v -X POST "http://localhost:5000?authorization=letmein"
> POST /?authorization=letmein HTTP/1.1
> Host: localhost:5000
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 202 Accepted
< Date: Wed, 15 May 2019 07:24:59 GMT
< Content-Length: 0
```

> MonoHook does not support HTTPS, therefore it is recommended that you put it
> behind a reverse proxy like Apache or nginx and use an SSL/TLS certificate.

### Concurrency

By default, MonoHook will allow only one execution of the command at a time. If
multiple requests are received in quick succession, they will be buffered until
previous executions have completed.

You may test this by running MonoHook with no options:

```
$> monohook -- echo Hook received
```

Then triggering 10 requests with cURL:

```
$> for n in 0 1 2 3 4 5 6 7 8 9; do curl -v -X POST http://localhost:5000; done
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
...
```

You can see in MonoHook's output that it is waiting for each command execution
to complete before starting the next one:

```
Executing /bin/echo
Hook received
Successfully executed /bin/echo

Executing /bin/echo
Hook received
Successfully executed /bin/echo

Executing /bin/echo
Hook received
Successfully executed /bin/echo

...
```

### TL;DR

```
monohook exposes a single HTTP webhook endpoint that executes a command.

Usage:
  monohook [OPTION...] [--] [EXEC...]

Options:
  -a, --authorization string   Authentication token that must be sent as a Bearer token in the 'Authorization' header or as the 'authorization' URL query parameter
  -b, --buffer uint            Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite) (default 10)
  -c, --concurrency uint       Maximum number of times the command should be executed in parallel (zero for infinite concurrency) (default 1)
  -C, --cwd string             Working directory in which to run the command
  -p, --port uint              Port on which to listen to (default 5000)
  -q, --quiet                  Do not print anything (default false)

Examples:
  Update a file when the hook is triggered:
    monohook -- touch hooked.txt
  Deploy an application when the hook is triggered:
    monohook -a letmein -- deploy-stuff.sh
```



## Exit codes

**MonoHook** may exit with the following status codes to indicate a problem:

Code | Description
:--- | :---
`1`  | Invalid arguments were given.
`2`  | The command to execute (provided after `--`) could not be found in the `$PATH`.
