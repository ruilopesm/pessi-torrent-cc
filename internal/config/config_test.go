package config

import (
	"os"
	"testing"
)

// TestNewConfig tests the NewConfig function
func TestNewConfig(t *testing.T) {
	// Create a temporary YAML file for testing
	testYAML := []byte(`
    dns:
      host: "10.4.4.2"
      port: 53

    tracker:
      host: "10.4.4.1"
      port: 42069

    node:
      port: 8081
  `)

	tempFile, err := os.CreateTemp("", "testconfig.yml")
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(testYAML); err != nil {
		t.Fatalf("Error writing to temporary file: %v", err)
	}

	// Test NewConfig function
	config, err := NewConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("Error creating config: %v", err)
	}

	// Verify the values in the config
	expectedConfig := &Config{
		DNS: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "10.4.4.2", Port: 53},
		Tracker: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "10.4.4.1", Port: 42069},
		Node: struct {
			Port int `yaml:"port"`
		}{Port: 8081},
	}

	// Compare the actual and expected config
	if *config != *expectedConfig {
		t.Errorf("Expected config: %+v, got: %+v", *expectedConfig, *config)
	}
}
