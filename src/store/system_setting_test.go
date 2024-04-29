package store

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"testing"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

const (
	latestSchemaFileName = "LATEST_SCHEMA.sql"
)

var d *sql.DB
var md *sql.DB
var dir string

// Initialize the logger and config
func init() {
	config.Opts = config.GetDefaultOptions()
	log.Logger = log.NewLogger()
}

func createDb() error {
	dir := os.TempDir()
	dir += "/e-oasis-test"
	// Create a directory if not exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}
	filename := dir + "/test_for_system_setting.db"
	metapath := dir + "/metadata.db"
	d, _ = sql.Open("sqlite", filename)
	md, _ = sql.Open("sqlite", metapath)
	return nil
}

//go:embed db/migration
var migrationFS embed.FS

func applyLatestSchema(db *sql.DB) error {
	// Read latest schema file
	latestSchemaPath := fmt.Sprintf("db/migration/%s", latestSchemaFileName)
	buf, err := migrationFS.ReadFile(latestSchemaPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to read latest schema file: %q", latestSchemaPath)
	}

	stmt := string(buf)
	if err := execute(stmt, db); err != nil {
		return errors.Wrapf(err, "Failed to apply latest schema: %s", stmt)
	}
	return nil
}

func execute(stmt string, d *sql.DB) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(stmt); err != nil {
		return errors.Wrap(err, "failed to execute statement")
	}

	return tx.Commit()
}

func TestGetOrUpsetSystemSetting(t *testing.T) {
	// Test code here
	if err := createDb(); err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	defer os.Remove(dir)
	applyLatestSchema(d)
	s := NewStore(d, md)
	system, err := s.GetOrUpsetSystemSecuritySetting()
	if err != nil {
		t.Fatalf("Failed to create system setting: %v", err)
	}
	t.Logf("System setting: %s", system.ToJSON())
	if system.JWTSecret == "" {
		t.Fatalf("JWT secret is empty")
	}
}

func TestGetBasicSystemSetting(t *testing.T) {
	// Test code here
	createDb()
	defer os.Remove(dir)
	applyLatestSchema(d)
	s := NewStore(d, md)
	basic, err := s.GetSystemBasicSetting()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to get system setting: %v", err)
		}
	}
	t.Logf("Basic system setting: %v", basic)
}

func TestGetGeneralSystemSetting(t *testing.T) {
	// Test code here
	createDb()
	defer os.Remove(dir)
	applyLatestSchema(d)
	s := NewStore(d, md)
	general, err := s.GetSystemGeneralSetting()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Failed to get system setting: %v", err)
		}
	}
	t.Logf("General system setting: %v", general)
}
