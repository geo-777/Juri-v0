package handlers

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RunHandler struct {
	executionService *services.ExecutionService
}

func NewRunHandler(service *services.ExecutionService) *RunHandler {
	return &RunHandler{
		executionService: service,
	}
}

func (h *RunHandler) Run(ctx *gin.Context) {

	var reqBody dtos.RunRequestDto

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	data, err := h.executionService.Run(&reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
