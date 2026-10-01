package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

	response, err := h.service.CreateSubmission(ctx.Request.Context(), reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	ctx.JSON(http.StatusAccepted, response)
}

func (h *HTTPHandler) GetSubmission(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	response, err := h.service.GetSubmission(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, response)
}
