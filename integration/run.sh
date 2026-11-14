#!/usr/bin/env bash

set -uo pipefail

currentdir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
workdir="$(mktemp -d)"
server_pid=""

cleanup() {
    if [ -n "$server_pid" ]; then
        kill "$server_pid" 2>/dev/null
        wait "$server_pid" 2>/dev/null
    fi
    rm -rf "$workdir"
}
trap cleanup EXIT

if [ $# -gt 0 ]; then
    read -ra ricol_binary <<< "$1"
else
    (cd "$currentdir/../src" && go build -o "$workdir/ricol" .) || exit 1
    ricol_binary=("$workdir/ricol")
fi

fail() {
    echo " -------- "
    echo "|  ERROR  |"
    echo " -------- "
    exit 1
}

check_output() {
    local status="$1" out="$2"
    if [ "$status" -ne 0 ] || grep -qi "error" <<< "$out"; then
        fail
    fi
}

run_program() {
    local ricol_file="$1" out status
    echo "\$ ricol integration/${ricol_file#"$currentdir/"}"
    out="$("${ricol_binary[@]}" "$ricol_file" < /dev/null 2>&1)"
    status=$?
    echo "$out"
    echo
    check_output "$status" "$out"
}

server_port() {
    grep -m 1 -oE '^server [A-Za-z_][A-Za-z0-9_]* at "[^"]*:[0-9]+"' "$1" | grep -oE '[0-9]+"$' | tr -d '"'
}

wait_for_port() {
    local port="$1"
    for _ in $(seq 50); do
        if ! kill -0 "$server_pid" 2>/dev/null; then
            return 1
        fi
        if (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then
            return 0
        fi
        sleep 0.1
    done
    return 1
}

run_server() {
    local ricol_file="$1" port="$2" name log status=0
    name="$(basename "$ricol_file" .ric)"
    log="$workdir/$name.log"
    echo "\$ ricol integration/$name.ric &"
    "${ricol_binary[@]}" "$ricol_file" < /dev/null > "$log" 2>&1 &
    server_pid=$!
    if ! wait_for_port "$port"; then
        cat "$log"
        fail
    fi
    echo
    for client_file in "$currentdir/clients/$name"*.ric; do
        if [ -e "$client_file" ]; then
            run_program "$client_file"
        fi
    done
    kill -0 "$server_pid" 2>/dev/null || status=1
    kill "$server_pid" 2>/dev/null
    wait "$server_pid" 2>/dev/null
    server_pid=""
    echo "(salida de integration/$name.ric)"
    cat "$log"
    echo
    check_output "$status" "$(cat "$log")"
}

for ricol_file in "$currentdir"/*.ric; do
    port="$(server_port "$ricol_file")"
    if [ -n "$port" ]; then
        run_server "$ricol_file" "$port"
    else
        run_program "$ricol_file"
    fi
done

echo " -------- "
echo "| Todo OK |"
echo " -------- "
