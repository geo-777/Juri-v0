# ADR-0005: Use Redis Queues for Asynchronous Judging

Status: Accepted

## Context

The judge endpoint currently performs compilation and testcase execution before
returning a response. A submission can take long enough to occupy an HTTP
request for its entire execution, and the API process is responsible for both
accepting requests and running jobs.

Juri already has a separate worker entry point. Moving execution to workers
allows the API to respond promptly and lets the number of workers scale
independently from the API.

## Options Considered

- Execute submissions synchronously in the API process
- Add a dedicated queue broker
- Use Redis as the job queue and job-result store

## Decision

Use Redis Streams consumer groups to queue judge jobs and distribute them to
workers.

The request flow is:

1. The API validates a judge request, assigns a unique job ID, and enqueues the
   job in Redis.
2. The API returns `202 Accepted` with the job ID and a URL for checking its
   status and result.
3. A worker reads a queued job, marks it as running, and evaluates it using
   the existing judge service.
4. The worker stores the terminal result in Redis, then acknowledges the
   stream entry.
5. The client polls the job URL to receive its current state and, when ready,
   the judge result.

Redis holds the queued request and job state/result. Jobs use at-least-once
delivery: workers acknowledge entries only after recording the result, and
pending entries can be reclaimed if a worker exits before completing a job.
Workers must tolerate redelivery by using the job ID to avoid creating
conflicting results. Job state and result records expire after a configured
retention period.

## Consequences

### Advantages

- HTTP requests return without waiting for compilation and execution
- API and worker capacity can scale independently
- Pending jobs can be recovered after a worker interruption
- Job IDs provide a stable handle for polling and result retrieval

### Disadvantages

- Adds Redis as an operational dependency
- Requires clients to handle polling and non-terminal job states
- At-least-once delivery requires idempotent worker behavior
- Results are available only for the configured Redis retention period unless
  durable result storage is added later
