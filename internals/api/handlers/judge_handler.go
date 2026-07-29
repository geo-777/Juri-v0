package handlers

import (
	"juri/internals/services"

	"github.com/gin-gonic/gin"
)

// JudgeHandler receives judge-related HTTP requests and forwards them to the service layer.
type JudgeHandler struct {
	service *services.JudgeService
}

func NewJudgeHandler(service *services.JudgeService) *JudgeHandler {
	return &JudgeHandler{
		service: service,
	}
}

// Judge handles judge requests for a submission.
func (h *JudgeHandler) Judge(ctx *gin.Context) {
	// TODO: implement the judge flow.
}
