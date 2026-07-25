package main

import (
	"juri/config"
	"juri/internals/api/handlers"
	"juri/internals/executor/compiler"
	"juri/internals/executor/runner"
	"juri/internals/services"
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
	comp := compiler.NewDockerCompiler(cfg, dockerClient)
	run := runner.NewDockerRunner(cfg, dockerClient)

	runService := services.NewRunService(comp, run)
	judgeService := services.NewJudgeService(comp, run)

	runHandler := handlers.NewRunHandler(runService)
	judgeHandler := handlers.NewJudgeHandler(judgeService)

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
	router.POST("/judge", judgeHandler.Judge)

	// Start Server
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
