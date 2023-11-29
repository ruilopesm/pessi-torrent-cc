package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "config/config.yml"
)

// Struct used by the YAML parser to parse the config file
type Config struct {
	DNS struct {
		Host string `yaml:"host"`
		Port uint   `yaml:"port"`
	} `yaml:"dns"`

	Tracker struct {
		Host string `yaml:"host"`
		Port uint   `yaml:"port"`
	} `yaml:"tracker"`

	Node struct {
		Port uint `yaml:"port"`
	} `yaml:"node"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
