package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"juri/internals/constants"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SubmissionService stores and retrieves submissions.
type SubmissionService struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
}

func NewSubmissionService(db *pgxpool.Pool, redisClient *redis.Client) *SubmissionService {
	return &SubmissionService{db: db, redisClient: redisClient}
}

func (s *SubmissionService) CreateSubmission(parent context.Context, dto SubmissionRequestDto) (*SubmissionResponseDTO, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	testCases, err := json.Marshal(dto.TestCases)
	if err != nil {
		return nil, fmt.Errorf("marshal submission test cases: %w", err)
	}
	timeLimit := dto.TimeLimitMs
	if timeLimit == 0 {
		timeLimit = constants.DefaultTimeLimitMs
	}
	memoryLimit := dto.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = constants.DefaultMemoryLimitKB
	}

	var id int64
	err = s.db.QueryRow(ctx, `
		INSERT INTO submissions (language, source_code, test_cases, callback_url, time_limit_ms, memory_limit_kb)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, dto.Language, dto.SourceCode, testCases, dto.CallbackURL, timeLimit, memoryLimit).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert submission: %w", err)
	}

	err = s.redisClient.RPush(ctx, constants.SubmissionQueueName, id).Err()
	if err != nil {
		return nil, fmt.Errorf("enqueue submission %d: %w", id, err)
	}

	return &SubmissionResponseDTO{ID: id, Status: constants.StatusPending}, nil
}

func (s *SubmissionService) GetSubmission(parent context.Context, id int64) (*JudgeResponseDto, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	var response JudgeResponseDto
	var status constants.Status
	var resultJSON []byte

	err := s.db.QueryRow(ctx, `
		SELECT id, status, COALESCE(stderr, ''), COALESCE(execution_time_ns, 0),
		       COALESCE(memory_kb, 0), COALESCE(result, '{}'::jsonb)
		FROM submissions
		WHERE id = $1
	`, id).Scan(
		&response.ID,
		&status,
		&response.Stderr,
		&response.ExecutionTimeNs,
		&response.MemoryKB,
		&resultJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("fetch submission: %w", err)
	}

	if status == constants.StatusQueued || status == constants.StatusRunning {
		response.Status = constants.StatusPending //coalscing both states for client
		return &response, nil
	}
	response.Status = status
	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &response.Result); err != nil {
			return nil, fmt.Errorf("decode submission result: %w", err)
		}
	}
	return &response, nil
}
