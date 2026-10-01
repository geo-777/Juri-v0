# Juri v0

Juri is a lightweight, standalone code-judging microservice designed to compile and execute submitted programs against test cases.

It is intended to serve as the execution and judging layer for coding platforms, handling tasks such as compilation, program execution, test-case evaluation, resource limiting, and result reporting. Juri is isolated from the main application so that code execution can be scaled and managed independently.
The current version is an early-stage implementation intended primarily for development, experimentation, and evaluation.

## Requirements

- Go 1.25.11 or later
- Docker, with its daemon running
- PostgreSQL
- Redis
- The `migrate` CLI

Install the PostgreSQL migration CLI if needed:

```sh
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Make sure Go's bin directory is on your `PATH`.

## Install and Run

1. Create a PostgreSQL database and start PostgreSQL and Redis.
2. Create a `.env` file in the project root with your local settings:

   ```dotenv
   PORT=8000
   WORKSPACE_ROOT=/relative/path/to/Juri-v0/temp-jobs
   REDIS_ADDRESS=localhost:6379
   REDIS_PASSWORD=
   DATABASE_URL=postgres://postgres:postgres@localhost:5432/juri?sslmode=disable
   ```

   Change the database URL to match your PostgreSQL user, password, and
   database. `WORKSPACE_ROOT` must be an absolute path writable by the worker.

3. Apply the database migrations:

   ```sh
   ./scripts/migrations.sh up
   ```

4. Build the language runtime images. Docker must be running:

   ```sh
   ./scripts/build-images.sh
   ```

5. In separate terminals, start the API and a worker from the project root:

   ```sh
   go run ./cmd/server
   ```

   ```sh
   go run ./cmd/worker
   ```

The API listens on port `8000` with the settings above. `GET /` is a basic
health check.

## Submit a Program

Create a submission:

```sh
curl -X POST http://localhost:8000/submissions \
  -H 'Content-Type: application/json' \
  -d '{"language":"python","source_code":"print(input())","test_cases":[{"input":"hello","expected_output":"hello\n"}]}'
```

The API returns `202 Accepted` with a submission ID. Check that submission for
its status and result:

```sh
curl http://localhost:8000/submissions/123
```

Replace `123` with the ID returned by the create request. A `callback_url` may
also be supplied to receive the completed result by HTTP POST.

## Current Features

- Asynchronous submissions using a Redis list and PostgreSQL persistence.
- Separate API and worker processes. Multiple instances can share the same
  PostgreSQL database and Redis queue; adding workers increases total job
  throughput. Each worker currently handles one job at a time.
- C, C++, Java 17, and Python 3.12 execution.
- Docker-based execution with network access disabled, a read-only container
  root filesystem, resource limits, and temporary workspaces.
- Compilation once per submission, followed by sequential test-case runs.
- Optional time and memory limits, plus optional completion callbacks.

## Planned

- Run multiple jobs concurrently within one worker.
- Reuse containers through a container pool.

## Current Limitations

This v0 queue does not guarantee delivery. A worker crash after removing a job
from Redis can leave that submission queued in PostgreSQL without an ID in the
Redis list. Do not treat the current setup as production-ready without
addressing recovery and operational needs.
