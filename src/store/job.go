package store

import (
	"github.com/Xunop/e-oasis/model"
)

func (s *Store) ListJobs() (*[]model.Job, error) {
	return nil, nil
}

func (s *Store) AddJob(job model.Job) (*model.Job, error) {
	stmt := `
    INSERT INTO job (user_id, path, type, status) VALUES (?, ?, ?, ?)
    RETURNING id, user_id, path, type, status
    `

    s.appDbLock.Lock()
    defer s.appDbLock.Unlock()
	tx, err := s.appDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var j model.Job
	if err := tx.QueryRow(stmt, job.UserID, job.Path, job.Type, job.Status).Scan(
		&j.ID, &j.UserID, &j.Path, &j.Type, &j.Status,
	); err != nil {
        return nil, err
    }
    if err := tx.Commit(); err != nil {
        return nil, err
    }

	return &j, nil
}
