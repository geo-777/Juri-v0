package handlers

import (
	"juri/internals/api/dtos"
	"juri/internals/services"
	"net/http"

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
	//binding and defining request body
	var reqBody dtos.JudgeRequestDto
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	response, err := h.service.Judge(reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	ctx.JSON(http.StatusOK, response)
}
