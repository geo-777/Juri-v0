# ADR-0002: Replace Docker CLI with Docker SDK

Status: Accepted

## Context

Early prototypes invoked Docker through shell commands.

This made error handling difficult and tightly coupled execution logic to the
Docker CLI.

## Decision

Use the official Docker Go SDK for all container operations.

Container lifecycle will be managed through the SDK.

## Consequences

### Advantages

- Structured API
- Better error handling
- No command parsing
- Easier testing
- Improved maintainability

### Disadvantages

- Larger dependency
- Slight learning curve
