package main

import (
	"code-runner/internals/api/handlers"
	"code-runner/internals/config"
	"code-runner/internals/executor"
	"code-runner/internals/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration:", err)
	}

	// Dependency Injection
	exec := executor.New(cfg)
	executionService := services.NewExecutionService(exec)
	runHandler := handlers.NewRunHandler(executionService)

	// Router setup
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Health Check
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"health": "ok",
		})
	})

	// Routes
	router.POST("/run", runHandler.Run)

	// Start Server
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
