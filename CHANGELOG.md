# Changelog

## v3.0.1

* Fix the `-p, --port` command-line flag which had no effect.

## v3.0.0

* **Breaking:** environment variables are no longer interpolated in command-line
  flag values.
* Environment variables can be used instead of command-line flags (e.g.
  `$MONOHOOK_CONCURRENCY` for `--concurrency`). Command-line flags take
  precedence over environment variables.
* Allow forwarding the HTTP request body, headers and URL to the command to
  execute.

## v2.0.0

* **Breaking:** require `POST` requests instead of accepting any method.
* Allow authorization to be provided with the `authorization` URL query
  parameter in addition to the `Authorization` header.
* Add a Dockerfile and [Docker Hub automated
  builds](https://hub.docker.com/r/alphahydrae/monohook).

## v1.0.0

* Initial release.
