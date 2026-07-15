package handlers

import (
	"code-runner/internals/api/dtos"
	"code-runner/internals/api/services"
	"code-runner/internals/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RunCodeHandlerFunc(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//reading body
		var reqBody dtos.RunRequestDto

		if err := ctx.ShouldBindJSON(&reqBody); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		err := services.RunService(&reqBody, cfg)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
