package db

import (
	"context"

	"github.com/Xunop/e-oasis/store"
)

func (d *DB) UpsertMigrationHistory(ctx context.Context, upsert *store.UpsertMigrationHistory) (*store.MigrationHistory, error) {
	stmt := `
		INSERT INTO migration_history (
			version
		)
		VALUES (?)
		ON CONFLICT(version) DO UPDATE
		SET
			version=EXCLUDED.version
		RETURNING version, created_ts
	`
	var migrationHistory store.MigrationHistory
	if err := d.DB.QueryRowContext(ctx, stmt, upsert.Version).Scan(
		&migrationHistory.Version,
		&migrationHistory.CreatedTs,
	); err != nil {
		return nil, err
	}

	return &migrationHistory, nil
}

func (d *DB) FindMigrationHistoryList(ctx context.Context, _ *store.FindMigrationHistory) ([]*store.MigrationHistory, error) {
	query := "SELECT `version`, `created_ts` FROM `migration_history` ORDER BY `created_ts` DESC"
	rows, err := d.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*store.MigrationHistory, 0)
	for rows.Next() {
		var mia store.MigrationHistory
		if err := rows.Scan(
			&mia.Version,
			&mia.CreatedTs,
		); err != nil {
			return nil, err
		}

		list = append(list, &mia)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
