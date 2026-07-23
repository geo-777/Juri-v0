#!/bin/sh

set +e

g++ main.cpp -std=c++20 -O2 -o main
if [ $? -ne 0 ]; then
cat > /workspace/metadata.txt <<EOF
runtime_ns=0
memory_kb=0
exit_code=1
phase=COMPILATION
EOF
    exit 1
fi

START=$(date +%s%N)

/usr/bin/time -q -f "%M" -o /tmp/memory.tmp ./main
EXIT_CODE=$?

END=$(date +%s%N)

cat > /workspace/metadata.txt <<EOF
runtime_ns=$((END - START))
memory_kb=$(cat /tmp/memory.tmp)
exit_code=$EXIT_CODE
phase=RUN
EOF

exit $EXIT_CODE