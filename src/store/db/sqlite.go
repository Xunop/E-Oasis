package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/pkg/errors"
	"modernc.org/sqlite"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/version"
	"github.com/Xunop/e-oasis/util"
)

type DB struct {
	*sql.DB
	name string
}

func init() {
	// Register custom functions
	sqlite.MustRegisterFunction("sortconcat", &sqlite.FunctionImpl{
		NArgs:         2,
		Deterministic: true,
		MakeAggregate: func(ctx sqlite.FunctionContext) (sqlite.AggregateFunction, error) {
			return util.NewSortedConcatenate(","), nil
		},
	})
	sqlite.MustRegisterFunction("concat", &sqlite.FunctionImpl{
		NArgs:         1,
		Deterministic: true,
		MakeAggregate: func(ctx sqlite.FunctionContext) (sqlite.AggregateFunction, error) {
			return util.NewConcatenate(","), nil
		},
	})
}

func NewDB(path, name string) (*DB, error) {
	if config.Opts.DSN == "" {
		return nil, errors.New("Database URL is required")
	}

	d, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db := &DB{d, name}

	return db, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

//go:embed migration
var migrationFS embed.FS

//go:embed seed
var seedFS embed.FS

// Migrate applies the latest schema to the database
func (d *DB) Migrate(ctx context.Context) error {
	switch d.name {
	case "system":
		fmt.Println("Migrate system database")
		return d.migrateSystem(ctx)
	case "meta":
		fmt.Println("Migrate meta database")
		return d.migrateMeta(ctx)
	default:
		return errors.New("Unknown db name")
	}
}

func (d *DB) migrateSystem(ctx context.Context) error {
	currentVersion := version.GetCurrentVersion()
	fmt.Println("Current version: ", currentVersion)
	if _, err := os.Stat(config.Opts.DSN); err != nil {
		// If the db file does not exist, create a new one with latest schema
		if errors.Is(err, os.ErrNotExist) {
			if err := d.applyLatestSchema(ctx); err != nil {
				return errors.Wrap(err, "failed to apply latest schema")
			}
			// Upset the newest version to migration table
			// Upsert the newest version to migration_history.
			if _, err := d.UpsertMigrationHistory(ctx, &store.UpsertMigrationHistory{
				Version: currentVersion,
			}); err != nil {
				return errors.Wrap(err, "failed to upsert migration history")
			}
		} else {
			return errors.Wrap(err, "failed to check database file")
		}
	} else {
		// If db file exist, check need to migrate or not
		migrationHistoryList, err := d.FindMigrationHistoryList(ctx, &store.FindMigrationHistory{})
		if err != nil {
			return errors.Wrap(err, "failed to find migration history list")
		}

		// If no migration history, apply latest schema
		if len(migrationHistoryList) == 0 {
			minorVersion := version.GetMinorVersion(currentVersion)
			if err := d.applyMigrationForMinorVersion(ctx, minorVersion); err != nil {
				return errors.Wrapf(err, "failed to apply version %s migration", minorVersion)
			}
			_, err = d.UpsertMigrationHistory(ctx, &store.UpsertMigrationHistory{
				Version: currentVersion,
			})
			if err != nil {
				return errors.Wrap(err, "failed to upsert migration history")
			}
			return nil
		}

		migrationHistoryVersionList := []string{}
		for _, migrationHistory := range migrationHistoryList {
			migrationHistoryVersionList = append(migrationHistoryVersionList, migrationHistory.Version)
		}
		// Sort and get the latest version
		slices.Sort(version.SortVersion(migrationHistoryVersionList))
		latestMigrationHistoryVersion := migrationHistoryVersionList[len(migrationHistoryVersionList)-1]

		if version.IsVersionGreaterThan(version.GetSchemaVersion(currentVersion), latestMigrationHistoryVersion) {
			minorVersionList := getMinorVersionList()
			// Backup the raw database file before migration
			rawBytes, err := os.ReadFile(config.Opts.DSN)
			if err != nil {
				return errors.Wrap(err, "failed to read raw database file")
			}
			backupDBFilePath := fmt.Sprintf("%s/e-oasis_%s_%d_backup.db", config.Opts.Data, version.GetCurrentVersion(), time.Now().Unix())
			if err := os.WriteFile(backupDBFilePath, rawBytes, 0644); err != nil {
				return errors.Wrap(err, "failed to write backup database file")
			}
			fmt.Println("Backup database file: ", backupDBFilePath)
			fmt.Printf("Start migration from %s to %s", latestMigrationHistoryVersion, currentVersion)
			for _, minorVersion := range minorVersionList {
				// Sicnce the patch don't need to update sql
				normalizedVersion := minorVersion + ".0"
				if version.IsVersionGreaterThan(normalizedVersion, latestMigrationHistoryVersion) && version.IsVersionGreaterOrEqualThan(currentVersion, normalizedVersion) {
					fmt.Println("Applying migration for", normalizedVersion)
					if err := d.applyMigrationForMinorVersion(ctx, minorVersion); err != nil {
						return errors.Wrap(err, "failed to apply minor version migration")
					}
				}
			}
			fmt.Println("End migrate")

			// Remove the created backup db file after migrate succeed.
			if err := os.Remove(backupDBFilePath); err != nil {
				fmt.Printf("Failed to remove temp database file, err: %v", err)
			}
		}
	}
	return nil
}

const (
	latestSystemSchemaFileName = "LATEST_SYSTEM_SCHEMA.sql"
	latestMetaSchemaFileName   = "LATEST_META_SCHEMA.sql"
	updateMetaSchemaFileName   = "UPDATE_META.sql"
)

// migrateMeta migrate the views in the calibre database
func (d *DB) migrateMeta(ctx context.Context) error {
	if _, err := os.Stat(config.Opts.MetaDSN); err != nil {
		// If the db file does not exist, create a new one with latest schema
		if errors.Is(err, os.ErrNotExist) {
			if err := d.applyLatestSchema(ctx); err != nil {
				return errors.Wrap(err, "failed to apply latest schema")
			}
		} else {
			return errors.Wrap(err, "failed to check database file")
		}
	} else {
		// If the db file exist, update the views.
		// Don't need to update the tables.
		exist, err := d.CheckTableExists(ctx, "migration_history")
		if err != nil {
			return errors.Wrap(err, "failed to check database table")
		}
		if exist {
			// Table exist and don't need to update
			return nil
		} else {
			// Table not exist, need to update
			// update will create `migration_history` table
			updatePath := fmt.Sprintf("migration/%s", updateMetaSchemaFileName)
			buf, err := migrationFS.ReadFile(updatePath)
			if err != nil {
				return errors.Wrapf(err, "failed to read update schema file: %q", updatePath)
			}

			stmt := string(buf)
			if err := d.execute(ctx, stmt); err != nil {
				return errors.Wrapf(err, "failed to apply update schema: %s", stmt)
			}
		}
	}
	return nil
}

func (d *DB) applyLatestSchema(ctx context.Context) error {
	// Read latest schema file
	latestSchemaPath := fmt.Sprintf("migration/%s", latestSystemSchemaFileName)
	if d.name == "meta" {
		latestSchemaPath = fmt.Sprintf("migration/%s", latestMetaSchemaFileName)
	}
	buf, err := migrationFS.ReadFile(latestSchemaPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read latest schema file: %q", latestSchemaPath)
	}

	stmt := string(buf)
	if err := d.execute(ctx, stmt); err != nil {
		return errors.Wrapf(err, "failed to apply latest schema: %s", stmt)
	}
	return nil
}

func (d *DB) applyMigrationForMinorVersion(ctx context.Context, minorVersion string) error {
	// Get all migration files for the minor version
	filenames, err := fs.Glob(migrationFS, fmt.Sprintf("migration/%s/*.sql", minorVersion))
	if err != nil {
		return errors.Wrapf(err, "Failed to find migration files for version %s", minorVersion)
	}

	// The filename files are sorted by name, so that they are applied in order.
	// 10001_example.sql, 10002_example.sql, 10003_example.sql, ...
	slices.Sort(filenames)
	migrationStmt := ""

	for _, filename := range filenames {
		buf, err := migrationFS.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "Failed to read migration file: %q", filename)
		}
		stmt := string(buf)
		migrationStmt += stmt
		if err := d.execute(ctx, migrationStmt); err != nil {
			return errors.Wrapf(err, "Failed to apply migration: %s", migrationStmt)
		}
	}

	// Upsert the newest version to migration_history.
	version := minorVersion + ".0"
	if _, err := d.UpsertMigrationHistory(ctx, &store.UpsertMigrationHistory{
		Version: version,
	}); err != nil {
		return errors.Wrapf(err, "Failed to upsert migration history for version %s", version)
	}

	return nil
}

func (d *DB) seed(ctx context.Context) error {
	filenames, err := fs.Glob(seedFS, fmt.Sprintf("%s/*.sql", "seed"))
	if err != nil {
		return errors.Wrap(err, "failed to read seed files")
	}

	slices.Sort(filenames)

	// Loop over all seed files and execute them in order.
	for _, filename := range filenames {
		buf, err := seedFS.ReadFile(filename)
		if err != nil {
			return errors.Wrapf(err, "Failed to read migration file: %q", filename)
		}
		stmt := string(buf)
		if err := d.execute(ctx, stmt); err != nil {
			return errors.Wrapf(err, "Seed error: %s", stmt)
		}
	}
	return nil
}

// execute runs a single SQL statement within a transaction.
func (d *DB) execute(ctx context.Context, stmt string) error {
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

// minorDirRegexp is a regular expression for minor version directory.
var minorDirRegexp = regexp.MustCompile(`^migration/[0-9]+\.[0-9]+$`)

func getMinorVersionList() []string {
	minorVersionList := []string{}

	if err := fs.WalkDir(migrationFS, "migration", func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() && minorDirRegexp.MatchString(path) {
			minorVersionList = append(minorVersionList, file.Name())
		}

		return nil
	}); err != nil {
		panic(err)
	}

	slices.Sort(version.SortVersion(minorVersionList))

	return minorVersionList
}
