package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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
	Opts.DSN = filepath.Join(Opts.Data, "/e-oasis.db")
	Opts.MetaDSN = filepath.Join(Opts.Data, "/metadata.db")
	fmt.Println("Data directory: ", Opts.Data)

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
		if !errors.Is(err, os.ErrNotExist) {
		    return "", errors.Wrapf(err, "unable to access data folder %s", dataDir)
		}
		// Create dir
		if dataDir == defaultData {
			err := os.MkdirAll(dataDir, 0755)
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					// Permission denied, try to create in user's home directory
					currentUser, err := user.Current()
					if err != nil {
						return "", errors.Wrap(err, "unable to get current user")
					}
					homeDir := currentUser.HomeDir
					fmt.Println("Permission denied, trying to check data folder in user's home directory")
					fmt.Printf("Home directory: %s\n", homeDir)

					if homeDir == "" {
						return "", errors.New("unable to get home directory")
					}

					// Check if data folder exists in user's home directory
					if _, err := os.Stat(filepath.Join(homeDir, "/.e-oasis")); err == nil {
						fmt.Println("Data folder exists in user's home directory: ", homeDir+"/.e-oasis")
						return filepath.Join(homeDir, "/.e-oasis"), nil
					}

					err = os.MkdirAll(filepath.Join(homeDir, "/.e-oasis"), 0755)
					if err != nil {
						return "", errors.Wrapf(err, "unable to create default data folder %s", dataDir)
					}
					fmt.Println("Data folder created in user's home directory: ", homeDir+"/.e-oasis")
					return filepath.Join(homeDir, "/.e-oasis"), nil
				}
				return "", errors.Wrapf(err, "unable to create default data folder %s", dataDir)
			}
		}
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
	err = viper.Unmarshal(Opts)
	if err != nil {
		return nil, err
	}
	return Opts, nil
}

// CheckSupportedTypes checks if the file type is supported
func CheckSupportedTypes(fileType string) bool {
	if len(Opts.SupportedTypes) == 0 {
		return false
	}

	for _, t := range Opts.SupportedTypes {
		if t == fileType {
			return true
		}
	}

	return false
}
