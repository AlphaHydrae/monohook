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

The command is executed asynchronously.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Installation](#installation)
  - [Docker](#docker)
  - [Download binary](#download-binary)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Concurrency](#concurrency)
  - [Buffering](#buffering)
  - [Command environment & working directory](#command-environment--working-directory)
  - [Forwarding HTTP request data to the command](#forwarding-http-request-data-to-the-command)
  - [Port number](#port-number)
  - [TL;DR](#tldr)
- [Exit codes](#exit-codes)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![version](https://img.shields.io/badge/Version-v2.0.0-blue.svg)](https://github.com/AlphaHydrae/monohook/releases/tag/v2.0.0)
[![build status](https://travis-ci.org/AlphaHydrae/monohook.svg?branch=master)](https://travis-ci.org/AlphaHydrae/monohook)
[![license](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.txt)



## Installation

### Docker

MonoHook is [available on Docker Hub][hub]. Run it with Docker:

```
$> docker run --rm -it -p 5000:5000 alphahydrae/monohook:2.0.0 -a letmein -- echo Hook triggered
```

Test it:

```
$> curl -v -X POST -H "Authorization: Bearer letmein" http://localhost:5000
```

You should see the command being executed in the container's logs:

```
Executing /bin/echo
Hook triggered
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
$> monohook -- echo Hook triggered
```

To trigger execution of the command, simply make a POST request to the hook:

```
$> curl -v -X POST http://localhost:5000
```

If you go back to the terminal where MonoHook is running, you will see that the
command's output is piped to MonoHook's own standard output so you can see it:

```
Executing /bin/echo
Hook triggered
Successfully executed /bin/echo
```

### Authentication

You can secure the webhook by requiring that an authentication token be sent,
using the `-a, --authorization <TOKEN>` command-line flag or the
`$MONOHOOK_AUTHORIZATION` environment variable:

```
$> monohook -a letmein -- echo Hook triggered
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
>
> For example, you could use this location in an nginx server block:
>
> ```
> location /hooks/deploy {
>   proxy_set_header X-Real-IP $remote_addr;
>   proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
>   proxy_set_header X-Forwarded-Proto https;
>   proxy_set_header Host $http_host;
>   proxy_redirect off;
>   proxy_pass http://localhost:3000;
> }
> ```

### Concurrency

By default, MonoHook will allow only one execution of the command at a time. If
multiple requests are received in quick succession, they will be buffered until
each previou execution has completed.

You may test this by running MonoHook with no options:

```
$> monohook -- echo Hook triggered
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
Hook triggered
Successfully executed /bin/echo

Executing /bin/echo
Hook triggered
Successfully executed /bin/echo

Executing /bin/echo
Hook triggered
Successfully executed /bin/echo

...
```

You may specify a greater concurrency with the `-c, --concurrency <LIMIT>`
command-line flag, or the `$MONOHOOK_CONCURRENCY` environment variable. This
example uses the `sleep` command to simulate a long-running command:

```
$> monohook -c 3 -- sleep 2
```

Trigger 10 requests like before:

```
$> for n in 0 1 2 3 4 5 6 7 8 9; do curl -v -X POST http://localhost:5000; done
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
...
```

This time, you will see MonoHook executing the command in parallel in batches of
3:

```
Executing /bin/sleep
Executing /bin/sleep
Executing /bin/sleep
Successfully executed /bin/sleep
Successfully executed /bin/sleep
Successfully executed /bin/sleep

Executing /bin/sleep
Executing /bin/sleep
Executing /bin/sleep
Successfully executed /bin/sleep
Successfully executed /bin/sleep
Successfully executed /bin/sleep

...
```

> You may specify infinite concurrency by setting a concurrency of zero:
> `monohook -c 0 -- sleep 2.

### Buffering

If concurrency is limited (which is the default) and MonoHook is already busy
executing a long-running command or commands, it will keep a buffer of pending
requests to execute. They will be executed as soon as the execution of previous
commands has completed and the configured concurrency limit permits it.

This buffer has a limited size (10 by default). If the buffer is full and
MonoHook receives more requests, it will discard them and respond with HTTP 429
Too Many Requests. Those requests will never be processed.

You may observe this behavior by running MonoHook with a small buffer and a
long-running command:

```
$> monohook -b 3 -- sleep 2
```

Trigger 10 requests like before:

```
$> for n in 0 1 2 3 4 5 6 7 8 9; do curl -v -X POST http://localhost:5000; done
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
< HTTP/1.1 202 Accepted
< HTTP/1.1 429 Too Many Requests
< HTTP/1.1 429 Too Many Requests
< HTTP/1.1 429 Too Many Requests
...
```

In this example, MonoHook immediately started executing the first request, then
put the next 3 in the buffer. Subsequent requests are discarded.

The size of the buffer can be configured with the `-b, --buffer <SIZE>`
command-line flag or the `$MONOHOOK_BUFFER` environment variable.

> You may specify an infinite buffer by setting the buffer to zero: `monohook -b
> 0 -- sleep 2`

### Command environment & working directory

The command will be executed with the same environment variables as those
available to MonoHook. By default, the working directory is the one where
MonoHook is run.

It can be configured to execute the command in a specific working directory with
the `-C, --cwd <DIR>` command-line flag or the `$MONOHOOK_CWD` environment
variable:

```
$> monohook -C /path/to/somewhere - pwd
```

### Forwarding HTTP request data to the command

The command does not have access to the HTTP request that triggered it by
default. You may choose to forward part or all of the following data:

* The HTTP request body with the `-B, --forward-request-body` command-line flag
  or the `$MONOHOOK_FORWARD_REQUEST_BODY` environment variable.

  If enabled, the request body will be piped to the command's standard input.
* HTTP request headers with the `-H, --forward-request-headers` command-line
  flag or the `$MONOHOOK_FORWARD_REQUEST_HEADERS` environment variable.

  If enabled, request headers will be provided to the command as environment
  variables, with the following naming convention:

      Accept -> $MONOHOOK_REQUEST_HEADER_ACCEPT
      Content-Type -> $MONOHOOK_REQUEST_HEADER_CONTENT_TYPE
* The HTTP request URL with the `-U, --forward-request-url` command-line flag or
  the `$MONOHOOK_FORWARD_REQUEST_URL` environment variable.

  If enabled, the URL parsed from the HTTP request line is provided to the
  command as the `$MONOHOOK_REQUEST_URL` environment variable.

The following example shows how to save the body of each HTTP request to a file:

```
$> monohook -B -- tee -a body.txt
```

Trigger 10 cURL requests with a body:

```
$> for n in 0 1 2 3 4 5 6 7 8 9; do curl -v -d $n http://localhost:5000; done
```

Check the file's contents:

```
$> cat body.txt
0123456789
```

> Note that forwarding the request body means that the execution of the command
> will not end before the HTTP request body has been completely consumed. This
> may have a performance impact. For example, with a concurrency of 1, each
> request will have to wait for the previous one to upload its data.

### Port number

MonoHook listens on port 5000 by default. You may use a different port with the
`-p, --port <NUMBER>` command-line flag or the `$MONOHOOK_PORT` environment
variable.

### TL;DR

```
monohook exposes a single HTTP webhook endpoint that executes a command.

Usage:
  monohook [OPTION...] [--] [EXEC...]

Options:
  -a, --authorization string      Authentication token that must be sent as a Bearer token in the 'Authorization' header or as the 'authorization' URL query parameter
  -b, --buffer uint               Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite) (default 10)
  -c, --concurrency uint          Maximum number of times the command should be executed in parallel (zero for infinite concurrency) (default 1)
  -C, --cwd string                Working directory in which to run the command
  -B, --forward-request-body      Whether to forward each HTTP request's body to the the command's standard input
  -H, --forward-request-headers   Whether to forward each HTTP request's headers to the the command as environment variables (e.g. Content-Type becomes $MONOHOOK_REQUEST_HEADER_CONTENT_TYPE)
  -U, --forward-request-url       Whether to forward each HTTP request's URL to the the command as the $MONOHOOK_REQUEST_URL environment variable
  -p, --port uint                 Port on which to listen to (default 5000)
  -q, --quiet                     Do not print anything except the command's standard output and error streams (default false)

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



[hub]: https://hub.docker.com/r/alphahydrae/monohook
