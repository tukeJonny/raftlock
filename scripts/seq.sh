#!/usr/bin/env bash

set -euo pipefail

if [[ $# != 1 ]]; then
    echo "Usage: ./scripts/seq.sh <number>"
    exit 1
fi

N=$1

for i in `seq 1 $N`; do
    ./bin/raftlock lock acquire $i
    echo "[+] Acquire $i"

    sleep 0.5

    ./bin/raftlock lock release $i
    echo "[-] Release $i"
done