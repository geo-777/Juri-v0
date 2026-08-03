# ADR-0004: Replace Docker Exec Per Testcase with a Persistent Runner

Status: Accepted

## Context

The original judge executed one Docker Exec operation per testcase.

For approximately 200 testcases this introduced around 20 seconds of overhead,
despite user programs requiring only milliseconds.

Profiling identified Docker Exec as the dominant cost.

## Decision

Replace repeated Docker Exec calls with a persistent runner process.

Execution flow:

Container Start

↓

Runner Starts

↓

Receive testcase

↓

Execute binary

↓

Return output

↓

Wait for next testcase

Container Stop

## Consequences

### Advantages

- Eliminates Docker Exec overhead
- Significantly lower latency
- Better scalability
- Closer to production online judges

### Disadvantages

- More complex runner protocol
- Stateful runner implementation
