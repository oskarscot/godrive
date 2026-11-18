package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func run() error {
	databseURL := os.Getenv("DATABASE_URL")

	if databseURL == "" {
		return fmt.Errorf("DATABAS_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databseURL)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v\n", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("Unabled to reach database: %v\n", err)
	}

	config := &config{
		address: ":8080",
	}

	application := &application{
		config: config,
		conn:   pool,
	}

	log.Printf("Starting server at: %s\n", config.address)

	return application.start(application.mount())
}

func main() {
	log.SetFlags(log.LstdFlags)

	if err := run(); err != nil {
		log.Printf("Failed to start app: %v", err)
		os.Exit(1)
	}
}
