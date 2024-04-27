package store // import "github.com/Xunop/e-oasis/store"

import (
	"database/sql"
	"sync"
)

type Store struct {
	db                 *sql.DB
	UserCache          sync.Map // map[int]*User
	UserSettingCache   sync.Map // map[string]*UserSetting
	SystemSettingCache sync.Map // map[string]*SystemSetting
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
