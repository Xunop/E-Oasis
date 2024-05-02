package store // import "github.com/Xunop/e-oasis/store"

import (
	"database/sql"
	"sync"
)

type Store struct {
	appDb              *sql.DB    // appDb is web server database
	appDbLock          sync.Mutex // appDbLock is used to lock appDb
	metaDb             *sql.DB    // metaDb is calibre database
	metaDbLock         sync.Mutex // metaDbLock is used to lock metaDb
	UserCache          sync.Map   // map[int]*User
	UserSettingCache   sync.Map   // map[string]*UserSetting
	SystemSettingCache sync.Map   // map[string]*SystemSetting
	JobCache           sync.Map   // map[int]*Job
	BookCache          sync.Map   // map[int]*Book
}

func NewStore(db *sql.DB, metaDB *sql.DB) *Store {
	return &Store{
		appDb:  db,
		metaDb: metaDB,
	}
}

func (s *Store) DBStats() sql.DBStats {
	return s.appDb.Stats()
}

func (s *Store) Ping() error {
	return s.appDb.Ping()
}

func (s *Store) Close() {
	//
}
