#!/bin/bash
# vim: ai:ts=8:sw=8:noet
set -eufCo pipefail
export SHELLOPTS
IFS=$'\t\n'

all_path="./tests/manual/manifests-all"
shuffle_path="./tests/manual/manifests-shuffle"
some_path="./tests/manual/manifests-some"
changed_path="./tests/manual/manifests-some-changed"
none_path="/dev/null"
config_file="./tests/manual/kahoy.yml"

function kahoy_apply() {
    [[ $# -ne 1 ]] && echo "USAGE: kahoy_apply NEW_MANIFESTS_PATH" && exit 1
    new_path="${1}"

    go run ./cmd/kahoy apply \
        --mode "paths" \
        --config-file "${config_file}" \
        --fs-old-manifests-path "${all_path}" \
        --fs-new-manifests-path "${new_path}"
}

case "${1:-}" in
"all")
    kahoy_apply "${all_path}"
    ;;
"shuffle")
    kahoy_apply "${shuffle_path}"
    ;;
"some")
    kahoy_apply "${some_path}"
    ;;
"changed")
    kahoy_apply "${changed_path}"
    ;;
"none")
    kahoy_apply "${none_path}" 
    ;;
*)
    echo "ERROR: Unknown command, use: 'all', 'shuffle', 'some', 'none' or 'changed'"
    exit 1
    ;;
esac
