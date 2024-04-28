package store

import (
	"database/sql"
	"encoding/json"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Store) GetSystemSetting(name string) (*model.SystemSetting, error) {
	if cache, ok := s.SystemSettingCache.Load(name); ok {
		return cache.(*model.SystemSetting), nil
	}

	// var setting model.SystemSetting
	setting := &model.SystemSetting{}
	stmt := `
    SELECT * FROM system_setting WHERE name = ?
	`
	if err := s.db.QueryRow(stmt, name).Scan(&setting.Name, &setting.Value, &setting.Description); err != nil {
		return nil, errors.Wrap(err, "failed to get system setting")
	}

	return setting, nil
}

func (s *Store) GetSystemBasicSetting() (*model.SystemSettingBasic, error) {
	systemSetting, err := s.GetSystemSetting(model.SettingTypeBasic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get system basic setting")
	}
	// var basicSetting model.SystemSettingBasic
	basicSetting := &model.SystemSettingBasic{}
	err = json.Unmarshal([]byte(systemSetting.Value), basicSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal system basic setting")
	}
	return basicSetting, nil
}

func (s *Store) GetSystemGeneralSetting() (*model.SystemSettingGeneral, error) {
	systemSetting, err := s.GetSystemSetting(model.SettingTypeGeneral)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get system general setting")
	}
	generalSetting := &model.SystemSettingGeneral{}
	err = json.Unmarshal([]byte(systemSetting.Value), generalSetting)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal system general setting")
	}
	return generalSetting, nil
}

// func (s *Store) GetSystemSecuritySetting() (*model.SystemSettingSecurity, error) {
// 	systemSetting, err := s.GetSystemSetting(model.SettingTypeSecurity)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// var securitySetting model.SystemSettingSecurity
// 	securitySetting := &model.SystemSettingSecurity{}
// 	err = json.Unmarshal([]byte(systemSetting.Value), securitySetting)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return securitySetting, nil
// }

func (s *Store) UpsetSystemSetting(setting *model.SystemSetting) (*model.SystemSetting, error) {
	newSetting := &model.SystemSetting{
		Name:        setting.Name,
		Description: setting.Description,
	}
	var value []byte
	var err error
	switch {
	case setting.Name == model.SettingTypeBasic:
		log.Debug("Setting type is basic")
		basic, err := setting.GetBasic()
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(basic)
	case setting.Name == model.SettingTypeSecurity:
		log.Debug("Setting type is security")
		security, err := setting.GetSecurity()
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(security)
	case setting.Name == model.SettingTypeGeneral:
		log.Debug("Setting type is general")
		general, err := setting.GetGeneral()
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(general)
	case setting.Name == model.SettingTypePlugins:
		log.Debug("Setting type is plugins")
		plugins, err := setting.GetPlugins()
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(plugins)
	case setting.Name == model.SettingTypeCustom:
		log.Debug("Setting type is custom")
		custom, err := setting.GetCustom()
		if err != nil {
			return nil, err
		}
		value, err = json.Marshal(custom)
	default:
		log.Debug("Unsupported workspace setting key", zap.String("setting", setting.Name))
		return nil, errors.Errorf("Unsupported workspace setting key: %v", setting.Name)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal/unmarshal setting value")
	}

	valueString := string(value)
	newSetting.Value = valueString
	stmt := `
	INSERT INTO system_setting (
		name, value, description
	)
	VALUES (?, ?, ?)
	ON CONFLICT(name) DO UPDATE
	SET
		value = EXCLUDED.value,
		description = EXCLUDED.description
	`
	if _, err := s.db.Exec(stmt, newSetting.Name, newSetting.Value, newSetting.Description); err != nil {
		return nil, errors.Wrap(err, "failed to insert/update system setting")
	}
	s.SystemSettingCache.Store(newSetting.Name, newSetting)
	return newSetting, nil
}

func (s *Store) GetOrUpsetSystemSecuritySetting() (*model.SystemSettingSecurity, error) {
	systemSetting, err := s.GetSystemSetting(model.SettingTypeSecurity)
	modified := false
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("SQL no rows, create security setting")
			modified = true
		} else {
			return nil, errors.Wrap(err, "failed to get security settings")
		}
	}
	if systemSetting == nil {
		log.Debug("No security setting found, create security setting")
		modified = true
	}

	securitySetting := &model.SystemSettingSecurity{}
	if systemSetting != nil {
		err = json.Unmarshal([]byte(systemSetting.Value), securitySetting)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal security settings")
		}
	}

	if securitySetting.JWTSecret == "" {
		log.Debug("No JWT secret found, create it")
		modified = true
	}

	log.Debug("Security setting status: ", zap.Bool("modified", modified))

	if modified {
		log.Debug("Start creating security setting")
		securitySetting := &model.SystemSettingSecurity{
			JWTSecret: util.GenUUID(),
		}
		_, err := s.UpsetSystemSetting(&model.SystemSetting{
			Name:  model.SettingTypeSecurity,
			Value: securitySetting.ToJSON(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to upset security settings")
		}
		log.Debug("Security setting created", zap.String("type", model.SettingTypeSecurity))
	}
	return securitySetting, nil
}

// TODO: Implement patch updates
func (s *Store) UpsetGeneralSettings(settings *model.SystemSettingGeneral) (*model.SystemSettingGeneral, error) {
	_, err := s.UpsetSystemSetting(&model.SystemSetting{
		Name:  model.SettingTypeGeneral,
		Value: settings.ToJSON(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to upset general settings")
	}

	return settings, nil
}
