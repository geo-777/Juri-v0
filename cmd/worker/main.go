package main

import (
	"context"
	"juri/config"
	"juri/internals/executor/workspace"
	"juri/internals/judge"
	"log"
	"strconv"
	"time"

	"github.com/moby/moby/client"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

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
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}

	// Wire the compiler, runner, and service layers together.
	runnerFactory := workspace.NewDockerRunnerFactory(cfg, dockerClient)
	judgeService := judge.NewJudgeService(runnerFactory)

	//redis listener

	for {
		result, err := redisClient.BRPop(ctx, 0, "submission_queue").Result()
		if err != nil {
			if ctx.Err() != nil {
				log.Println("worker shutting down")
				return
			}

			log.Printf("redis BRPOP error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		submissionID, err := strconv.Atoi(result[1])
		if err != nil {
			log.Printf("invalid submission ID: %v", err)
			continue
		}

		// process submissionID
		log.Printf("Processing submissionID : %d", submissionID)
	}

}
