#!/bin/sh

set -e

gcc main.c -std=c11 -O2 -o main

START=$(date +%s%N)

/usr/bin/time -f "__MEMORY_KB__=%M" ./main

END=$(date +%s%N)

RUNTIME=$((END - START))

echo "__RUNTIME_NS__=$RUNTIME"