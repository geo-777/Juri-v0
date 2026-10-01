package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"juri/internals/api"
	"juri/internals/constants"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Judge evaluates a submission using the configured execution backend.
type Judge interface {
	Judge(context.Context, api.SubmissionRequestDto) (*api.JudgeResponseDto, error)
}

// Worker loads queued submissions, runs them, and stores their result.
type Worker struct {
	db         *pgxpool.Pool
	judge      Judge
	httpclient *http.Client
}

func New(db *pgxpool.Pool, judge Judge) *Worker {
	return &Worker{db: db, judge: judge, httpclient: &http.Client{Timeout: 10 * time.Second}}
}

// Process claims and evaluates one submission. Already claimed or completed
// submissions are harmless no-ops, which makes duplicate queue IDs safe.
func (w *Worker) Process(ctx context.Context, id int64) error {
	// Claim only queued submissions; duplicate or completed IDs are no-ops.
	submission, callbackURL, claimed, err := w.claimSubmission(ctx, id)
	if err != nil {
		if claimed {
			return w.finishSystemError(ctx, id, callbackURL, err)
		}
		return err
	}
	if !claimed {
		return nil
	}

	result, err := w.evaluateSubmission(ctx, submission)
	if err != nil {
		return w.finishSystemError(ctx, id, callbackURL, err)
	}
	result.ID = id
	// Save the result before notifying the client callback.
	if err := w.storeResult(ctx, id, result); err != nil {
		return fmt.Errorf("store result for submission %d: %w", id, err)
	}

	w.notifyCallback(id, callbackURL, result)
	return nil
}

func (w *Worker) claimSubmission(ctx context.Context, id int64) (api.SubmissionRequestDto, string, bool, error) {
	var submission api.SubmissionRequestDto
	var callbackURL string
	var casesJSON []byte

	// The status check makes claiming a submission atomic.
	err := w.db.QueryRow(ctx, `
		UPDATE submissions
		SET status = 'running', updated_at = NOW()
		WHERE id = $1 AND status = 'queued'
		RETURNING language, source_code, test_cases, COALESCE(callback_url, ''), time_limit_ms, memory_limit_kb
	`, id).Scan(&submission.Language, &submission.SourceCode, &casesJSON, &callbackURL, &submission.TimeLimitMs, &submission.MemoryLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return submission, "", false, nil
	}
	if err != nil {
		return submission, "", false, fmt.Errorf("claim submission %d: %w", id, err)
	}
	if err := json.Unmarshal(casesJSON, &submission.TestCases); err != nil {
		return submission, callbackURL, true, fmt.Errorf("decode test cases: %w", err)
	}

	return submission, callbackURL, true, nil
}

func (w *Worker) evaluateSubmission(ctx context.Context, submission api.SubmissionRequestDto) (*api.JudgeResponseDto, error) {
	result, err := w.judge.Judge(ctx, submission)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("judge returned an empty result")
	}
	return result, nil
}

func (w *Worker) notifyCallback(id int64, callbackURL string, result *api.JudgeResponseDto) {
	if callbackURL == "" {
		return
	}
	// Callback failures are logged and do not undo the saved result.
	if err := w.postCallback(callbackURL, result); err != nil {
		log.Printf("callback for submission %d failed: %v", id, err)
	}
}

func (w *Worker) storeResult(ctx context.Context, id int64, result *api.JudgeResponseDto) error {
	resultJSON, err := json.Marshal(result.Result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	// Only update submissions that this worker successfully claimed.
	_, err = w.db.Exec(ctx, `
		UPDATE submissions
		SET status = $2, stderr = $3, execution_time_ns = $4, memory_kb = $5,
		    result = $6, updated_at = NOW()
		WHERE id = $1 AND status = 'running'
	`, id, result.Status, result.Stderr, result.ExecutionTimeNs, result.MemoryKB, resultJSON)
	return err
}

func (w *Worker) finishSystemError(ctx context.Context, id int64, callbackURL string, cause error) error {
	_, err := w.db.Exec(ctx, `
		UPDATE submissions
		SET status = 'system_error', stderr = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'running'
	`, id, cause.Error())
	if err != nil {
		return fmt.Errorf("judge submission %d: %v; persist system error: %w", id, cause, err)
	}

	log.Printf("submission %d failed internally: %v", id, cause)
	if callbackURL != "" {
		result := &api.JudgeResponseDto{ID: id, Status: constants.StatusSystemError, Stderr: cause.Error()}
		if err := w.postCallback(callbackURL, result); err != nil {
			log.Printf("callback for submission %d failed: %v", id, err)
		}
	}
	return nil
}

func (w *Worker) postCallback(url string, result *api.JudgeResponseDto) error {
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode callback: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpclient.Do(req)
	if err != nil {
		return fmt.Errorf("send callback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("callback returned HTTP %d", resp.StatusCode)
	}
	return nil
}
