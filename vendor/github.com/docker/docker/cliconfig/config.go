package cliconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/cliconfig/configfile"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/engine-api/types"
)

const (
	// ConfigFileName is the name of config file
	ConfigFileName = "config.json"
	configFileDir  = ".docker"
	oldConfigfile  = ".dockercfg"

	// This constant is only used for really old config files when the
	// URL wasn't saved as part of the config file and it was just
	// assumed to be this value.
	defaultIndexserver = "https://index.docker.io/v1/"
)

var (
	configDir = os.Getenv("DOCKER_CONFIG")
)

func init() {
	if configDir == "" {
		configDir = filepath.Join(homedir.Get(), configFileDir)
	}
}

// ConfigDir returns the directory the configuration file is stored in
func ConfigDir() string {
	return configDir
}

// SetConfigDir sets the directory the configuration file is stored in
func SetConfigDir(dir string) {
	configDir = dir
}

// NewConfigFile initializes an empty configuration file for the given filename 'fn'
func NewConfigFile(fn string) *configfile.ConfigFile {
	return &configfile.ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
		HTTPHeaders: make(map[string]string),
		Filename:    fn,
	}
}

// LegacyLoadFromReader is a convenience function that creates a ConfigFile object from
// a non-nested reader
func LegacyLoadFromReader(configData io.Reader) (*configfile.ConfigFile, error) {
	configFile := configfile.ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
	}
	err := configFile.LegacyLoadFromReader(configData)
	return &configFile, err
}

// LoadFromReader is a convenience function that creates a ConfigFile object from
// a reader
func LoadFromReader(configData io.Reader) (*configfile.ConfigFile, error) {
	configFile := configfile.ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
	}
	err := configFile.LoadFromReader(configData)
	return &configFile, err
}

// Load reads the configuration files in the given directory, and sets up
// the auth config information and returns values.
// FIXME: use the internal golang config parser
func Load(configDir string) (*configfile.ConfigFile, error) {
	if configDir == "" {
		configDir = ConfigDir()
	}

	configFile := configfile.ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
		Filename:    filepath.Join(configDir, ConfigFileName),
	}

	// Try happy path first - latest config file
	if _, err := os.Stat(configFile.Filename); err == nil {
		file, err := os.Open(configFile.Filename)
		if err != nil {
			return &configFile, fmt.Errorf("%s - %v", configFile.Filename, err)
		}
		defer file.Close()
		err = configFile.LoadFromReader(file)
		if err != nil {
			err = fmt.Errorf("%s - %v", configFile.Filename, err)
		}
		return &configFile, err
	} else if !os.IsNotExist(err) {
		// if file is there but we can't stat it for any reason other
		// than it doesn't exist then stop
		return &configFile, fmt.Errorf("%s - %v", configFile.Filename, err)
	}

	// Can't find latest config file so check for the old one
	confFile := filepath.Join(homedir.Get(), oldConfigfile)
	if _, err := os.Stat(confFile); err != nil {
		return &configFile, nil //missing file is not an error
	}
	file, err := os.Open(confFile)
	if err != nil {
		return &configFile, fmt.Errorf("%s - %v", confFile, err)
	}
	defer file.Close()
	err = configFile.LegacyLoadFromReader(file)
	if err != nil {
		return &configFile, fmt.Errorf("%s - %v", confFile, err)
	}

	if configFile.HTTPHeaders == nil {
		configFile.HTTPHeaders = map[string]string{}
	}
	return &configFile, nil
}
