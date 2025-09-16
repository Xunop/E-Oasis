package model //import "github.com/Xunop/e-oasis/model"

import "encoding/json"

const (
	SettingTypeBasic    = "SETTINGS_BASIC"
	SettingTypeGeneral  = "SETTINGS_GENERAL"
	SettingTypePlugins  = "SETTINGS_PLUGINS"
	SettingTypeSecurity = "SETTINGS_SECURITY"
	SettingTypeCustom   = "SETTINGS_CUSTOM"
)

type SystemSetting struct {
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

type SystemSettingBasic struct {
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`
}

func (s *SystemSettingBasic) ToJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type SystemSettingGeneral struct {
	DisableSignup         bool `json:"disallow_registration"`
	DisallowPasswordLogin bool `json:"disallow_password_login"`
}

func (s *SystemSettingGeneral) ToJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type SystemSettingPlugins struct {
}

func (s *SystemSettingPlugins) ToJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type SystemSettingSecurity struct {
	JWTSecret string `json:"jwt_secret,omitempty"`
}

type SystemSettingCustom struct {
	Settings map[string]interface{} `json:"settings"`
}

func (s *SystemSettingSecurity) ToJSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *SystemSetting) GetBasic() (*SystemSettingBasic, error) {
	var basic SystemSettingBasic
	err := json.Unmarshal([]byte(s.Value), &basic)
	if err != nil {
		return nil, err
	}
	return &basic, nil
}

func (s *SystemSetting) GetGeneral() (*SystemSettingGeneral, error) {
	var general SystemSettingGeneral
	err := json.Unmarshal([]byte(s.Value), &general)
	if err != nil {
		return nil, err
	}
	return &general, nil
}

func (s *SystemSetting) GetPlugins() (*SystemSettingPlugins, error) {
	var plugins SystemSettingPlugins
	err := json.Unmarshal([]byte(s.Value), &plugins)
	if err != nil {
		return nil, err
	}
	return &plugins, nil
}

func (s *SystemSetting) GetSecurity() (*SystemSettingSecurity, error) {
	var security SystemSettingSecurity
	err := json.Unmarshal([]byte(s.Value), &security)
	if err != nil {
		return nil, err
	}
	return &security, nil
}

func (s *SystemSetting) GetCustom() (*SystemSettingCustom, error) {
	var custom SystemSettingCustom
	err := json.Unmarshal([]byte(s.Value), &custom)
	if err != nil {
		return nil, err
	}
	return &custom, nil
}
