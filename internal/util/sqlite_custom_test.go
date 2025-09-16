package util

import (
	"database/sql"
	"testing"

	sqlite3 "modernc.org/sqlite"
)

func registger() {
	sqlite3.MustRegisterFunction("sortconcat", &sqlite3.FunctionImpl{
		NArgs:         2,
		Deterministic: true,
		MakeAggregate: func(ctx sqlite3.FunctionContext) (sqlite3.AggregateFunction, error) {
			return NewSortedConcatenate(","), nil
		},
	})
	sqlite3.MustRegisterFunction("concat", &sqlite3.FunctionImpl{
		NArgs:         1,
		Deterministic: true,
		MakeAggregate: func(ctx sqlite3.FunctionContext) (sqlite3.AggregateFunction, error) {
			return NewConcatenate(","), nil
		},
	})
}

func TestCustomFunction(t *testing.T) {
	registger()
	withDB := func(test func(db *sql.DB)) {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()
		test(db)
	}

	t.Run("Test SortedConcatenate", func(tt *testing.T) {
		withDB(func(db *sql.DB) {
			if _, err := db.Exec("DROP TABLE IF EXISTS test; CREATE TABLE IF NOT EXISTS test (id INTEGER, value TEXT); INSERT INTO test VALUES (1, '一'), (3, '二'), (2, '三'), (4, 'four')"); err != nil {
				tt.Errorf("Error: %v", err)
			}
			row := db.QueryRow("SELECT sortconcat(id, value) FROM test")

			var result string
			if err := row.Scan(&result); err != nil {
				tt.Errorf("Error: %v", err)
			}
			tt.Log("result: ", result)
			if result != "一,三,二,four" {
				tt.Errorf("Expected: %s, got: %s", "一,三,二,four", result)
			}
		})
	})

	t.Run("Test Concatenate", func(tt *testing.T) {
		withDB(func(db *sql.DB) {
			if _, err := db.Exec("DROP TABLE IF EXISTS test; CREATE TABLE IF NOT EXISTS test (id INTEGER, value TEXT); INSERT INTO test VALUES (1, '一'), (3, '二'), (2, '三'), (4, 'four')"); err != nil {
				tt.Errorf("Error: %v", err)
			}
			row := db.QueryRow("SELECT concat(value) FROM test")

			var result string
			if err := row.Scan(&result); err != nil {
				tt.Errorf("Error: %v", err)
			}
			tt.Log("result: ", result)
			if result != "一,二,三,four" {
				tt.Errorf("Expected: %s, got: %s", "一,二,三,four", result)
			}
		})
	})
}
