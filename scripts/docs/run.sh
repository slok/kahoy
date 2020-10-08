#!/bin/bash
# vim: ai:ts=8:sw=8:noet
set -eufCo pipefail
export SHELLOPTS
IFS=$'\t\n'

command -v hugo >/dev/null 2>&1 || { echo 'please install hugo'; exit 1; }

hugo server -s ./docs-src --bind=0.0.0.0 