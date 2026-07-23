#!/bin/sh

set +e

javac Main.java

START=$(date +%s%N)

/usr/bin/time -q -f "%M" -o /tmp/memory.tmp java Main
EXIT_CODE=$?

END=$(date +%s%N)

RUNTIME=$((END - START))

cat > /workspace/metadata.txt <<EOF
runtime_ns=$RUNTIME
memory_kb=$(cat /tmp/memory.tmp)
exit_code=$EXIT_CODE
EOF

exit $EXIT_CODE