package services

import (
	"juri/internals/api/dtos"
)

// JudgeService is responsible for evaluating submissions against judge criteria.
type JudgeService struct {
}

func NewJudgeService() *JudgeService {
	return &JudgeService{}
}

// Judge is the entry point for judge-related evaluation flow.
func (s *JudgeService) Judge(dto dtos.JudgeRequestDto) (*dtos.JudgeResponseDto, error) {
	return nil, nil
}
