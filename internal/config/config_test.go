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
		Host: %s
		Port: %d
		DSN: %s
		LogLevel: %s
		Data: %s
		`, opts.Host, opts.Port, opts.DSN, opts.LogLevel, opts.Data)

	if opts.Data != "/var/opt/e-oasis" {
		t.Errorf("data not set")
	}
}

func TestLoadConfigFile(t *testing.T) {
    opts, err := ParseFile("config_test.toml")
    if err != nil {
        t.Errorf("Error loading config: %s", err)
    }
	t.Logf(`Config
		Host: %s
		Port: %d
		DSN: %s
		LogLevel: %s
		LogFile: %s
		`, opts.Host, opts.Port, opts.DSN, opts.LogLevel, opts.LogFile)
	if opts.Host != "127.0.0.1" {
		t.Errorf("host incorrect")
	}
	if opts.LogFile != "test.log" {
		t.Errorf("log_file incorrect")
	}
	if opts.Port != 2333 {
		t.Errorf("port incorrect")
	}
	if opts.LogLevel != "debug" {
		t.Errorf("log_level incorrect")
	}
}
