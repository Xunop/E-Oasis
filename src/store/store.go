package store

import (
	"database/sql"
	"sync"
)

type Store struct {
	db                 *sql.DB
	userCache          sync.Map // map[int]*User
	userSettingCache   sync.Map // map[string]*UserSetting
	systemSettingCache sync.Map // map[string]*SystemSetting
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) DBStats() sql.DBStats {
	return s.db.Stats()
}

func (s *Store) Ping() error {
	return s.db.Ping()
}

func (s *Store) Close() {
	//
}
