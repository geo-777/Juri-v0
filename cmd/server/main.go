package main

import (
	"juri/config"
	"juri/internals/api/handlers"
	"juri/internals/executor/workspace"

	"juri/internals/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

func main() {
	// Initialize the Docker client used by the execution backends.
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	// Load runtime settings from the environment.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration:", err)
	}

	// Wire the compiler, runner, and service layers together.
	runnerFactory := workspace.NewDockerRunnerFactory(cfg, dockerClient)

	runService := services.NewRunService(runnerFactory)
	judgeService := services.NewJudgeService(runnerFactory)

	runHandler := handlers.NewRunHandler(runService)
	judgeHandler := handlers.NewJudgeHandler(judgeService)

	// Configure the Gin router and register the public endpoints.
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Expose a basic health endpoint for readiness checks.
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"health": "ok",
		})
	})

	// Register the submission and judging routes.
	router.POST("/run", runHandler.Run)
	router.POST("/judge", judgeHandler.Judge)

	// Start the HTTP server.
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
