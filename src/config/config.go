package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Xunop/e-oasis/version"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var Opts *Options

func GetConfig() (*Options, error) {
	GetDefaultOptions()

	dataDir, err := checkDataDir(Opts.Data)
	if err != nil {
		fmt.Println("Error checking data directory: ", err)
		return nil, err
	}

	Opts.Data = dataDir
	if Opts.DSN == "" {
		dbFile := filepath.Join(Opts.Data, "e-oasis.db")
		Opts.DSN = dbFile
	}

	Opts.Version = version.GetCurrentVersion()

	return Opts, nil
}

func checkDataDir(dataDir string) (string, error) {
	// Convert to absolute path if relative path is supplied.
	if !filepath.IsAbs(dataDir) {
		relativeDir := filepath.Join(filepath.Dir(os.Args[0]), dataDir)
		absDir, err := filepath.Abs(relativeDir)
		if err != nil {
			return "", err
		}
		dataDir = absDir
	}

	// Trim trailing \ or / in case user supplies
	dataDir = strings.TrimRight(dataDir, "\\/")
	if _, err := os.Stat(dataDir); err != nil {
		// Create dir
		if dataDir == defaultData {
			err := os.MkdirAll(dataDir, 0755)
			if err != nil {
				return "", errors.Wrapf(err, "unable to create default data folder %s", dataDir)
			}
		}
		return "", errors.Wrapf(err, "unable to access data folder %s", dataDir)
	}
	return dataDir, nil
}

func ParseFile(file string) (*Options, error) {
	// Check if file exists
	if _, err := os.Stat(file); err != nil {
		return nil, errors.Wrapf(err, "unable to access config file %s", file)
	}

	viper.SetConfigFile(file)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	Opts = &Options{}
	err = viper.Unmarshal(Opts)
	if err != nil {
		return nil, err
	}
	return Opts, nil
}
