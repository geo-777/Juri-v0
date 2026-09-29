package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	service *SubmissionService
}

func NewHTTPHandler(service *SubmissionService) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

func (h *HTTPHandler) CreateSubmission(ctx *gin.Context) {
	var reqBody SubmissionRequestDto
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	response, err := h.service.CreateSubmission(reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	ctx.JSON(http.StatusOK, response)
}
