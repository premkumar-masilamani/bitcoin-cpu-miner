package config

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bitcoin Bitcoin `yaml:"bitcoin"`
}

type Bitcoin struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
}

func LoadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file in %v", file)
	}

	appConfig := &Config{}
	if err = yaml.Unmarshal(data, appConfig); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config file in %v", file)
	}
	log.Println("Loaded the config from file: ", file)
	return appConfig, nil
}
