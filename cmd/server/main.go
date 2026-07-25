package main

import (
	"juri/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

func main() {
	//init docker client
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration:", err)
	}

	// Dependency Injection

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
	//router.POST("/run", runHandler.Run)

	// Start Server
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
