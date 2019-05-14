# monohook

Run a single HTTP webhook endpoint that executes a command (e.g. a deployment script).

```
$> monohook --authorization letmein --concurrency 1 --port 3000 -- deploy.sh
```

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![version](https://img.shields.io/badge/Version-v2.0.0-blue.svg)](https://github.com/AlphaHydrae/monohook/releases/tag/v2.0.0)
[![build status](https://travis-ci.org/AlphaHydrae/monohook.svg?branch=master)](https://travis-ci.org/AlphaHydrae/monohook)
[![license](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.txt)



## Installation

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

```
monohook runs a single HTTP webhook endpoint that executes a command.

Usage:
  monohook [OPTION...] [--] [EXEC...]

Options:
  -a, --authorization string   Authentication token that must be sent as a Bearer token in the 'Authorization' header or as the 'authorization' URL query parameter
  -b, --buffer string          Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite) (default "10")
  -c, --concurrency string     Maximum number of times the command should be executed in parallel (zero for infinite concurrency) (default "1")
  -C, --cwd string             Working directory in which to run the command
  -p, --port string            Port on which to listen to (default "5000")
  -q, --quiet                  Do not print anything (default false)

Examples:
  Update a file when the hook is triggered:
    monohook -- touch hooked.txt
  Deploy an application when the hook is triggered:
    monohook -a letmein -- deploy-stuff.sh
```



## Exit codes

**monohook** may exit with the following status codes:

Code | Description
:--- | :---
`1`  | Invalid arguments were given.
`2`  | An unrecoverable error occurred while trying to interpolate environment variables into options or command arguments.
`3`  | The command to execute (provided after `--`) could not be found in the `$PATH`.
