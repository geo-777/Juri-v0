#!/bin/sh

set -e

START=$(date +%s%N)

/usr/bin/time -f "__MEMORY_KB__=%M" python3 main.py

END=$(date +%s%N)

RUNTIME=$((END - START))

echo "__RUNTIME_NS__=$RUNTIME"