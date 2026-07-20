#!/bin/sh

set -e

g++ main.cpp -o main

START=$(date +%s%N)

./main

END=$(date +%s%N)

RUNTIME=$((END - START))

echo "__RUNTIME_NS__=$RUNTIME"