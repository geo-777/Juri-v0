
CREATE TYPE submission_language AS ENUM (
    'cpp',
    'c',
    'java',
    'python'
);

CREATE TYPE submission_status AS ENUM (
    'queued',
    'running',
    'success',
    'wrong_answer',
    'compilation_error',
    'runtime_error',
    'time_limit_exceeded',
    'memory_limit_exceeded',
    'system_error'
);

CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    language submission_language NOT NULL,
    test_cases JSONB NOT NULL,
    callback_url TEXT,
    time_limit_ms INT NOT NULL DEFAULT 2000,
    memory_limit_kb INT NOT NULL DEFAULT 131072,
    status submission_status NOT NULL DEFAULT 'queued',
    source_code TEXT NOT NULL DEFAULT '',
    stderr TEXT,
    execution_time_ns BIGINT,
    memory_kb BIGINT,
    result JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE submissions
ADD CONSTRAINT submissions_time_limit_check
CHECK (time_limit_ms > 0 AND time_limit_ms <= 15000);

ALTER TABLE submissions
ADD CONSTRAINT submissions_memory_limit_check
CHECK (memory_limit_kb > 0 AND memory_limit_kb <= 1048576);

--for operational and debugging queries
CREATE INDEX idx_submissions_status
ON submissions(status);

CREATE INDEX idx_submissions_created_at
ON submissions(created_at DESC);