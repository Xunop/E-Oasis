package store

import (
	"database/sql"
	"os"
	"testing"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
)

var testBookDb *sql.DB
var testBookMetaDb *sql.DB
var testBookDir string

// Initialize the logger and config
func init() {
	config.Opts = config.GetDefaultOptions()
	log.Logger = log.NewLogger()
}

func createBookTestDb() error {
	testBookDir = os.TempDir()
	testBookDir += "/e-oasis-test"
	// Create a directory if not exists
	if _, err := os.Stat(testBookDir); os.IsNotExist(err) {
		err := os.Mkdir(testBookDir, 0755)
		if err != nil {
			return err
		}
	}
	filename := testBookDir + "/test_for_test_book.db"
	metapath := testBookDir + "/metadata.db"
	testBookDb, _ = sql.Open("sqlite", filename)
	testBookMetaDb, _ = sql.Open("sqlite", metapath)
	return nil
}

func TestSetBookStatus(t *testing.T) {
	if createBookTestDb() != nil {
		t.Fatalf("Failed to create database")
	}
	// defer os.RemoveAll(testBookDir)
	applyLatestSchema(testBookDb)

	s := NewStore(testBookDb, testBookMetaDb)
	_, err := s.CreateUser(&model.User{
		ID:           1,
		Username:     "test",
		PasswordHash: "test",
		Role:         model.RoleUser,
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var bookStatus = &model.BookReadingStatusLink{
		ID:         1,
		BookID:     1,
		UserID:     1, // Must exist in the database
		Status:     1,
		Percentage: 1,
		Page:       10,
		CurPage:    5,
		LastRead:   "2021-10-10",
	}

	newStatus, err := s.SetBookStatus(bookStatus)
	if err != nil {
		t.Fatalf("Failed to set book status: %v", err)
	}
	if newStatus == nil {
		t.Fatalf("Failed to set book status")
	}

	bookStatus = &model.BookReadingStatusLink{
		ID:         1,
		BookID:     1,
		UserID:     1, // Must exist in the database
		Status:     1,
		Percentage: 2,
		Page:       10,
		CurPage:    10,
		LastRead:   "2021-12-10",
	}

	newStatus, err = s.SetBookStatus(bookStatus)
	if err != nil {
		t.Fatalf("Failed to set book status: %v", err)
	}
	if newStatus == nil {
		t.Fatalf("Failed to set book status")
	}
}
