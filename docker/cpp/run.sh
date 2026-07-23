#!/bin/sh

set +e

g++ main.cpp -std=c++20 -O2 -o main

START=$(date +%s%N)

/usr/bin/time -q -f "%M" -o /tmp/memory.tmp ./main
EXIT_CODE=$?

END=$(date +%s%N)

RUNTIME=$((END - START))

cat > /workspace/metadata.txt <<EOF
runtime_ns=$RUNTIME
memory_kb=$(cat /tmp/memory.tmp)
exit_code=$EXIT_CODE
EOF

exit $EXIT_CODE