package main

import (
	"code-runner/internals/api/handlers"
	"code-runner/internals/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	//loading env configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load env files.")
		return
	}

	//router
	var router *gin.Engine = gin.Default()
	router.SetTrustedProxies(nil)
	//health route
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"Health":   "Ok",
			"Database": "Connected",
		})
	})

	//runner route
	router.POST("/run", handlers.RunCodeHandlerFunc(cfg))
	router.Run(":" + cfg.Port)
}
