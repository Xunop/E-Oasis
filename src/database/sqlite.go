package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/version"
	_ "modernc.org/sqlite"
)

type DB sql.DB

func NewDB() (*sql.DB, error) {
	if config.Opts.DSN == "" {
		return nil, errors.New("Database URL is required")
	}

	// TODO: Add parameter
	db, err := sql.Open("sqlite", config.Opts.DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//go:embed migration
var migrationFS embed.FS

//go:embed seed
var seedFS embed.FS

// TODO:
func Migrate(db *sql.DB, ctx context.Context) error {
	currentVersion := version.GetCurrentVersion()
	fmt.Println("Current version: ", currentVersion)
	if _, err := os.Stat(config.Opts.DSN); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := applyLatestSchema(db); err != nil {
				return errors.Wrap(err, "Failed to apply latest schema")
			}
			// Upset the newest version to migration table
		} else {
			return errors.Wrap(err, "Failed to check database file")
		}
	}
	return nil
}

const (
	latestSchemaFileName = "LATEST__SCHEMA.sql"
)

func applyLatestSchema(db *sql.DB) error {
	// Read latest schema file
	latestSchemaPath := fmt.Sprintf("migration/%s", latestSchemaFileName)
	buf, err := migrationFS.ReadFile(latestSchemaPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to read latest schema file: %q", latestSchemaPath)
	}

	stmt := string(buf)
	if err := execute(db, context.Background(), stmt); err != nil {
		return errors.Wrapf(err, "Failed to apply latest schema: %s", stmt)
	}
	return nil
}

// execute runs a single SQL statement within a transaction.
func execute(d *sql.DB, ctx context.Context, stmt string) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, stmt); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return tx.Commit()
}
