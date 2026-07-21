#!/bin/sh

set +e

START=$(date +%s%N)

/usr/bin/time -f "%M" -o /tmp/memory.tmp python3 main.py
EXIT_CODE=$?

END=$(date +%s%N)

RUNTIME=$((END - START))

cat > /workspace/metadata.txt <<EOF
runtime_ns=$RUNTIME
memory_kb=$(cat /tmp/memory.tmp)
exit_code=$EXIT_CODE
EOF

exit $EXIT_CODE