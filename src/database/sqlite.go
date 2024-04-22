package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Xunop/e-oasis/config"
	_ "modernc.org/sqlite"
)

func NewDB(config *config.Config) (*sql.DB, error) {
	if config.DsnURI == "" {
		return nil, errors.New("Database URL is required")
	}

	db, err := sql.Open("sqlite", config.DsnURI)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB, ctx context.Context) error {
	return nil
}
