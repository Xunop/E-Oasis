package config

import (
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
    opts, err := GetConfig()
    if err != nil {
        t.Errorf("Error loading config: %s", err)
    }

	t.Logf(`Config
		Version: %s
		Host: %s
		Port: %d
		DSN: %s
		LogLevel: %s
		Data: %s
		`, opts.Version, opts.Host, opts.Port, opts.DSN, opts.LogLevel, opts.Data)

	if opts.Version != defaultVersion {
		t.Errorf("Version not set")
	}
}

func TestLoadConfigFile(t *testing.T) {
    opts, err := ParseFile("config_test.toml")
    if err != nil {
        t.Errorf("Error loading config: %s", err)
    }
	t.Logf(`Config
		Version: %s
		Host: %s
		Port: %d
		DSN: %s
		LogLevel: %s
		LogFile: %s
		`, opts.Version, opts.Host, opts.Port, opts.DSN, opts.LogLevel, opts.LogFile)
    if opts.Version != "1.0.0" {
		t.Errorf("Version not set")
	}
	if opts.Host != "127.0.0.1" {
		t.Errorf("Host not set")
	}
	if opts.LogFile != "test.log" {
		t.Errorf("LogFile not set")
	}
	if opts.Port != 2333 {
		t.Errorf("Port not set")
	}
	if opts.LogLevel != "DEBUG" {
		t.Errorf("LogLevel not set")
	}
}
