package handlers

import (
	"juri/internals/api/dtos"
	"juri/internals/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RunHandler struct {
	service *services.RunService
}

func NewRunHandler(service *services.RunService) *RunHandler {
	return &RunHandler{
		service: service,
	}
}

func (h *RunHandler) Run(ctx *gin.Context) {

	var req dtos.RunRequestDto

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	response, err := h.service.Run()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
