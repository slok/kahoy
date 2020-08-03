#!/usr/bin/env sh

set -o errexit
set -o nounset

go test -race -tags='integration' ./test/integration/...