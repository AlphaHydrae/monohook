SHELL := /usr/bin/env bash

test:
	go test -count=1 -v ./...

install:
	dep ensure

build:
	./scripts/build.sh

doctoc:
	command -v doctoc &>/dev/null && doctoc README.md || { >&2 echo "Error: install doctoc with \`npm install -g doctoc\`"; exit 1; }
