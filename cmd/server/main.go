package main

import (
	"context"
	"juri/config"
	"juri/internals/api"
	"juri/internals/executor/workspace"
	"juri/internals/judge"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
	"github.com/redis/go-redis/v9"
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

	//init redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}

	// Wire the compiler, runner, and service layers together.
	runnerFactory := workspace.NewDockerRunnerFactory(cfg, dockerClient)

	_ = judge.NewJudgeService(runnerFactory)

	// Wire up submission service and handler
	submissionSvc := api.NewSubmissionService(redisClient)
	submissionHandler := api.NewHTTPHandler(submissionSvc)

	// Configure the Gin router and register the public endpoints.
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Expose a basic health endpoint for readiness checks.
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"health": "ok",
		})
	})

	//submission route
	router.POST("/submissions", submissionHandler.CreateSubmission)

	// Start the HTTP server.
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
