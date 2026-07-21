#!/bin/sh

set -e

g++ main.cpp -std=c++20 -O2 -o main

START=$(date +%s%N)

/usr/bin/time -f "__MEMORY_KB__=%M" ./main

END=$(date +%s%N)

RUNTIME=$((END - START))

echo "__RUNTIME_NS__=$RUNTIME"