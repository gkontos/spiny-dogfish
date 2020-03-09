package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// AppConfig is the top level config
type AppConfig struct {
	App Application `toml:"app"`
}

// Application contains configurations for the app
type Application struct {
	ProjectRoot           string `toml:"project_root"`
	ExternalConfiguration string `toml:"external_properties"`
}

// LoadAppConfig will load configs from a toml config file
func LoadAppConfig(file string) (*AppConfig, error) {
	conf := &AppConfig{}
	if file == "" {
		return conf, nil
	}
	tomlData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %q: %v", file, err)
	}

	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		return nil, fmt.Errorf("could not read config file %q: %v", file, err)
	}
	return conf, nil
}
