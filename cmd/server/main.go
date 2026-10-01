package main

import (
	"context"
	"juri/config"
	"juri/internals/api"
	"juri/pkg/database"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load runtime settings from the environment.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration:", err)
	}

	//setting up db connection
	pool, err := database.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed.")
		return
	}
	defer pool.Close()

	//init redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}

	// Wire up submission service and handler
	submissionSvc := api.NewSubmissionService(pool, redisClient)
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

	// Submission routes
	router.POST("/submissions", submissionHandler.CreateSubmission)
	router.GET("/submissions/:id", submissionHandler.GetSubmission)

	// Start the HTTP server.
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
