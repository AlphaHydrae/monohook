#!/usr/bin/env bash
set -e

bold=$(tput bold)
normal=$(tput sgr0)
versions="darwin_amd64 linux_amd64 linux_arm64 windows_amd64"

test -z "$RELEASE" && RELEASE="$(git describe --tags || echo -n)"

if test -z "$RELEASE"; then
  >&2 echo No Git tag found
  exit 1
fi

rm -fr release/${RELEASE}

printf "\n${bold}○ Building binaries...${normal}\n"

for version in $versions; do
  os="$(echo $version | cut -d _ -f 1)"
  arch="$(echo $version | cut -d _ -f 2)"
  env GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o release/${RELEASE}/monohook_${os}_${arch} &
done

wait

for version in $versions; do
  file="release/${RELEASE}/monohook_${version}"
  printf "\n${bold}○ Compressing ${file}...${normal}\n\n"
  upx --ultra-brute "$file"
done

printf "\n${bold}○ Calculating digests...${normal}\n\n"
dgstore 'release/**/*'

echo
