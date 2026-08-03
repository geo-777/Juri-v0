# ADR-0003: Separate Compilation and Execution

Status: Accepted

## Context

Compilation and execution are fundamentally different stages.

Compilation should happen once while execution may happen multiple times
against different test cases.

## Decision

Split the execution pipeline into two independent stages.

Compilation:

- Compile source
- Produce executable
- Return compilation diagnostics

Execution:

- Execute compiled artifact
- Provide stdin
- Capture stdout/stderr
- Record execution metrics

## Consequences

### Advantages

- Avoid repeated compilation
- Simpler execution flow
- Supports judge endpoints naturally
- Easier future caching

### Disadvantages

- More components
- Slightly more coordination
