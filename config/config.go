package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type LogConfigs struct {
	ApplicationLog LogConfig `toml:"applicationLog"`
	AccessLog      LogConfig `toml:"accessLog"`
}

type LogConfig struct {
	Output   string `toml:"output"`
	Level    string `toml:"level"`
	Filename string `toml:"filename"`
}

type AppConfig struct {
	App Application `toml:"app"`
}

type Application struct {
	ProjectRoot string `toml:"project_root"`
}

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

func LoadLogConfig(file string) (*LogConfigs, error) {
	conf := &LogConfigs{}
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
