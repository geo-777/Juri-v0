# ADR-0001: Use Docker as the Initial Sandbox

Status: Accepted

## Context

Juri requires an isolated environment for executing untrusted user code safely.
The sandbox should provide filesystem isolation, resource limits, and network
isolation while remaining easy to install and contribute to.

## Options Considered

- Docker
- nsjail
- isolate
- Firecracker microVMs

## Decision

Use Docker as the initial sandbox implementation.

Each submission executes inside a dedicated Docker container with:

- Network disabled
- Memory limits
- CPU limits
- Non-root execution
- Ephemeral workspace

## Consequences

### Advantages

- Mature ecosystem
- Excellent Go SDK
- Cross-platform
- Easy contributor onboarding
- Widely understood deployment model

### Disadvantages

- Higher process startup overhead
- Docker daemon dependency
- Slightly slower than lightweight sandboxes

## Future Work

Introduce additional sandbox implementations (e.g. nsjail) behind a common
interface if performance becomes a bottleneck.
