package api

import (
	"github.com/redis/go-redis/v9"
)

// SubmissionService is responsible for creating jobs (asynchronous approach)
type SubmissionService struct {
	redisClient *redis.Client
}

func NewSubmissionService(redisClient *redis.Client) *SubmissionService {
	return &SubmissionService{redisClient: redisClient}
}

func (s *SubmissionService) CreateSubmission(dto SubmissionRequestDto) (*JudgeResponseDto, error) {
	return nil, nil
}
