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
        --provider "paths" \
        --config-file "${config_file}" \
        --fs-old-manifests-path "${all_path}" \
        --fs-new-manifests-path "${new_path}" \
        --auto-approve \
        --report-path=-
}

function kahoy_kube_apply() {
    [[ $# -ne 1 ]] && echo "USAGE: kahoy_apply NEW_MANIFESTS_PATH" && exit 1
    new_path="${1}"

    go run ./cmd/kahoy apply \
        --provider "kubernetes" \
        --config-file "${config_file}" \
        --fs-new-manifests-path "${new_path}" \
        --auto-approve \
        --kube-provider-id "kahoy-test" \
        --kube-provider-namespace "default" \
        --report-path=-
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
"kube-all")
    kahoy_kube_apply "${all_path}"
    ;;
"kube-shuffle")
    kahoy_kube_apply "${shuffle_path}"
    ;;
"kube-some")
    kahoy_kube_apply "${some_path}"
    ;;
"kube-changed")
    kahoy_kube_apply "${changed_path}"
    ;;
"kube-none")
    kahoy_kube_apply "${none_path}" 
    ;;
*)
    echo "ERROR: Unknown command, use: 'all', 'shuffle', 'some', 'none', 'kube-changed', 'kube-all', 'kube-shuffle', 'kube-some', 'kube-none' or 'kube-changed'"
    exit 1
    ;;
esac
