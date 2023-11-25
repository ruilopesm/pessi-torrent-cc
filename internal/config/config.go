package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

const (
	ConfigPath = "./config/config.yml"
)

type Config struct {
	DNS struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"dns"`

	Tracker struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"tracker"`

	Node struct {
		Port int `yaml:"port"`
	} `yaml:"node"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
