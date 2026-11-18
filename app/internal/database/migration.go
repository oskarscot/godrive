package database

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}

func Migrate(pgx *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(pgx, dir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
