#!/usr/bin/env bash

set -euo pipefail

currentdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
output="$currentdir/results/measurements.txt"
builddir="$(mktemp -d)"
trap 'rm -rf "$builddir"' EXIT

mkdir -p "$currentdir/results"
: > "$output"
(cd "$currentdir/../src" && go build -o "$builddir/ricol" .)

for benchdir in "$currentdir"/*/; do
    name="$(basename "$benchdir")"
    [ "$name" == "results" ] && continue

    gcc -O2 -o "$builddir/$name-c" "$benchdir/$name.c"
    go build -o "$builddir/$name-go" "$benchdir/$name.go"

    commands=(
        "$builddir/$name-c"
        "$builddir/$name-go"
        "python3 $benchdir/$name.py"
        "$builddir/ricol $benchdir/$name.ric"
    )

    expected="$($builddir/$name-c)"
    for command in "${commands[@]}"; do
        if [ "$($command)" != "$expected" ]; then
            echo "The output of '$command' does not match the output of C ($expected)" >&2
            exit 1
        fi
    done

    echo "\$ benchmark $name" | tee -a "$output"
    hyperfine --shell=none --style=basic --warmup 2 --runs 10 \
        -n C "${commands[0]}" \
        -n Go "${commands[1]}" \
        -n Python "${commands[2]}" \
        -n Ricol "${commands[3]}" \
        | tee -a "$output"
done

echo "Measurements saved to benchmarks/results/measurements.txt"
