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

    row := s.db.QueryRow(stmt, job.UserID, job.Path, job.Type, job.Status)

    var j model.Job
    err := row.Scan(&j.ID, &j.UserID, &j.Path, &j.Type, &j.Status)
    if err != nil {
        return nil, err
    }

    s.JobCache.Store(j.ID, &j)
    return &j, nil
}
