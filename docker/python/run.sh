#!/bin/sh

set -e

START=$(date +%s%N)

python3 main.py

END=$(date +%s%N)

RUNTIME=$((END - START))

MEMORY=$(grep "Maximum resident set size" memory.txt | awk '{print $6}')

echo "__RUNTIME_NS__=$RUNTIME"
echo "__MEMORY_KB__=$MEMORY"