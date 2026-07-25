package handlers

import (
	"juri/internals/services"

	"github.com/gin-gonic/gin"
)

type JudgeHandler struct {
	service *services.JudgeService
}

func NewJudgeHandler(service *services.JudgeService) *JudgeHandler {
	return &JudgeHandler{
		service: service,
	}
}

func (h *JudgeHandler) Judge(ctx *gin.Context) {

}
