#!/bin/sh

set -e

javac Main.java

START=$(date +%s%N)

java Main

END=$(date +%s%N)

RUNTIME=$((END - START))

echo "__RUNTIME_NS__=$RUNTIME"