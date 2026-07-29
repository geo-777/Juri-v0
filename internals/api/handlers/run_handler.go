package handlers

import (
	"juri/internals/api/dtos"
	"juri/internals/services"
	"juri/internals/utils"
	"log"
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

	var reqBody dtos.RunRequestDto

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	response, err := h.service.Run(reqBody)
	if err != nil {
		requestID := utils.GenerateNanoId(10)
		log.Printf("request_id=%s endpoint=/run error=%v", requestID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":      "internal server error",
			"request_id": requestID,
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
