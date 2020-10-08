#!/bin/bash
# vim: ai:ts=8:sw=8:noet
set -eufCo pipefail
export SHELLOPTS
IFS=$'\t\n'

command -v hugo >/dev/null 2>&1 || { echo 'please install hugo'; exit 1; }

# Clean.
rm -rf ./docs && mkdir ./docs

# Create Github's custom CNAME.
echo "docs.kahoy.dev" > ./docs/CNAME

# Generate.
hugo -s ./docs-src