package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/oskarscot/godrive/internal/database"
	"github.com/oskarscot/godrive/migrations"
)

func run() error {
	logger := log.Default()

	databseURL := os.Getenv("DATABASE_URL")

	if databseURL == "" {
		return fmt.Errorf("DATABAS_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		logger: logger,
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	logger.Println("Checking migrations...")
	if err = database.MigrateFS(db, migrations.FS, "."); err != nil {
		return fmt.Errorf("Failed to migrate: %w", err)
	}

	logger.Printf("Starting server at: %s\n", config.address)

	handler, err := application.mount(ctx)
	if err != nil {
		return err
	}

	return application.start(handler)
}

func main() {
	log.SetFlags(log.LstdFlags)

	if err := run(); err != nil {
		log.Printf("Failed to start app: %v\n", err)
		os.Exit(1)
	}
}
