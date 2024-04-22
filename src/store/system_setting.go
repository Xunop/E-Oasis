package store

import (
	"github.com/Xunop/e-oasis/model"
)

func (s *Store) GetSystemSetting(name string) (model.SystemSetting, error) {
	// var setting SystemSetting
	// s.db.QueryRow("SELECT * FROM system_settings WHERE name = $1", name).Scan(&setting.Name, &setting.Value, &setting.Description)
	return model.SystemSetting{}, nil
}
