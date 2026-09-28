#!/usr/bin/env bash

set -uo pipefail

currentdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ $# -gt 0 ]; then
    read -ra ricol_binary <<< "$1"
else
    builddir="$(mktemp -d)"
    trap 'rm -rf "$builddir"' EXIT
    (cd "$currentdir/../src" && go build -o "$builddir/ricol" .) || exit 1
    ricol_binary=("$builddir/ricol")
fi

for ricol_file in "$currentdir"/*.ric; do
    echo "\$ ricol integration/$(basename "$ricol_file")"

    out="$("${ricol_binary[@]}" "$ricol_file" < /dev/null 2>&1)"
    status=$?
    echo "$out"
    echo

    if [ $status -ne 0 ] || grep -qi "error" <<< "$out"; then
        echo " -------- "
        echo "|  ERROR  |"
        echo " -------- "
        exit 1
    fi
done

echo " -------- "
echo "| Todo OK |"
echo " -------- "
