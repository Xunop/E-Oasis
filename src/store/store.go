package store

import (
	"database/sql"
	"sync"

	"github.com/Xunop/e-oasis/config"
)

type Store struct {
	config           *config.Config
	db               *sql.DB
	userCache        sync.Map // map[int]*User
	userSettingCache sync.Map // map[string]*UserSetting
}

func NewStore(db *sql.DB, config *config.Config) *Store {
	return &Store{
		db:     db,
		config: config,
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
