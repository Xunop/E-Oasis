package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Xunop/e-oasis/config"
	_ "modernc.org/sqlite"
)

func NewDB() (*sql.DB, error) {
	if config.Opts.DSN == "" {
		return nil, errors.New("Database URL is required")
	}

	db, err := sql.Open("sqlite", config.Opts.DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TODO:
func Migrate(db *sql.DB, ctx context.Context) error {
	return nil
}
