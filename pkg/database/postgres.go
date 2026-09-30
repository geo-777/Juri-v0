package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDatabase(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	//converts databaseURL to pgxpool.Config structure format
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Printf("Parsing config from DatabaseURL resulted in error : %v", err)
		return nil, err
	}
	//creating pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Database connection Error : %v", err)
		return nil, err
	}
	//ping test
	if err = pool.Ping(ctx); err != nil {
		log.Printf("Pinging not successful : %v", err)
		pool.Close()
		return nil, err
	}

	return pool, nil
}
