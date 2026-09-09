package handlers

import (
	"juri/internals/api/dtos"
	"juri/internals/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RunHandler receives HTTP requests for submission execution and delegates them to the service layer.
type RunHandler struct {
	service *services.RunService
}

func NewRunHandler(service *services.RunService) *RunHandler {
	return &RunHandler{
		service: service,
	}
}

func (h *RunHandler) Run(ctx *gin.Context) {
	// Parse and validate the incoming JSON payload.
	var reqBody dtos.RunRequestDto

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Send the request to the domain service and return the result.
	response, err := h.service.Run(reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error :" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
