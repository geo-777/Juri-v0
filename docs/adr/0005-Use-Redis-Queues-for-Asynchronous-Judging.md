# ADR-0005: Use a Redis List for Asynchronous Judging

Status: Accepted

## Context

Compilation and testcase execution can take longer than an HTTP request should
remain open. The API and judge worker therefore need a way to hand off
submissions asynchronously.

## Decision

The API saves each submission in PostgreSQL, pushes its ID onto the Redis list
`submission_queue`, and returns `202 Accepted` with the ID and `pending` status.
Workers block on the list, claim queued submissions in PostgreSQL, run the
judge, and save the final status and result back to PostgreSQL. Clients can
check progress and results through the submission endpoint. If a callback URL
was provided, the worker posts the result to it after saving.

Redis carries submission IDs only. PostgreSQL is the source of truth for
submission data, status, and results.

## Consequences

- API requests return without waiting for judging to finish.
- API and worker processes can scale independently.
- Redis is an additional runtime dependency.
- `BRPOP` removes an ID as soon as a worker receives it. The worker requeues an
  ID when processing returns an error, but a worker crash after removal can
  leave a submission queued in PostgreSQL with no queue entry. This setup does
  not provide automatic recovery or guaranteed delivery.
