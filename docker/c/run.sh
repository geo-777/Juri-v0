#!/bin/sh

set +e

gcc main.c -O2 -o main
COMPILE_EXIT=$?

if [ $COMPILE_EXIT -ne 0 ]; then
cat > /workspace/metadata.txt <<EOF
runtime_ns=0
memory_kb=0
exit_code=$COMPILE_EXIT
phase=COMPILATION
EOF

    exit $COMPILE_EXIT
fi

START=$(date +%s%N)

/usr/bin/time -q -f "%M" -o /tmp/memory.tmp ./main
EXIT_CODE=$?

END=$(date +%s%N)

RUNTIME=$((END - START))

cat > /workspace/metadata.txt <<EOF
runtime_ns=$RUNTIME
memory_kb=$(cat /tmp/memory.tmp)
exit_code=$EXIT_CODE
phase=RUN
EOF

exit $EXIT_CODE