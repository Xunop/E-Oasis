package store // import "github.com/Xunop/e-oasis/store"

import (
	"database/sql"
	"sync"
)

type Store struct {
	db                 *sql.DB
	metaDb             *sql.DB  // metaDb is calibre database
	UserCache          sync.Map // map[int]*User
	UserSettingCache   sync.Map // map[string]*UserSetting
	SystemSettingCache sync.Map // map[string]*SystemSetting
	JobCache           sync.Map // map[int]*Job
	BookCache          sync.Map // map[int]*Book
}

func NewStore(db *sql.DB, metaDB *sql.DB) *Store {
	return &Store{
		db:     db,
		metaDb: metaDB,
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
