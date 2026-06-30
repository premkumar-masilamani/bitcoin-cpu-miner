package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
bitcoin:
  host: "127.0.0.1"
  port: "18443"
  username: "bitcoinrpcuser"
  password: "hard-to-guess-rpc-password"
  address: "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We"
`
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "127.0.0.1", cfg.Bitcoin.Host)
	assert.Equal(t, "18443", cfg.Bitcoin.Port)
	assert.Equal(t, "bitcoinrpcuser", cfg.Bitcoin.Username)
	assert.Equal(t, "hard-to-guess-rpc-password", cfg.Bitcoin.Password)
	assert.Equal(t, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", cfg.Bitcoin.Address)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create a temp config file with invalid YAML
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
bitcoin:
  host: [invalid: value
`
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config file")
}
