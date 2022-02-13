#!/usr/bin/env bash

set -eo pipefail

version=$((cat version/version.go | grep 'Version =' | sed -n 's/^.*\([0-9]\+\.[0-9]\+\.[0-9]\+-gitlab\.[0-9]\+\).*$/\1/p') || echo dev2 | sed -e 's/^v//g')
exact_tag=$(git describe --exact-match --tags --always 2>/dev/null | sed -e 's/^v//g' || echo "")

if [[ $(echo ${exact_tag} | grep -E "^[0-9]+\.[0-9]+\.[0-9]+-gitlab\.[0-9]+(-rc[0-9]+)?$") ]]; then
    echo ${exact_tag}
    exit 0
fi

pre_release_info=$(git describe --always --long | sed -r "s/v[0-9\.]+-gitlab\.[0-9]+(-rc[0-9]+)?-//")

echo "${version}-beta-${pre_release_info}"

