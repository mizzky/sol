package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadDBPoolConfig_Defaults(t *testing.T) {
	getenv := func(string) string {
		return ""
	}

	got, err := loadDBPoolConfig(getenv)

	require.NoError(t, err)
	require.Equal(t, 10, got.maxOpenConns)
	require.Equal(t, 5, got.maxIdleConns)
	require.Equal(t, 30*time.Minute, got.connMaxLifetime)
}

func TestLoadDBPoolConfig_Overrides(t *testing.T) {
	env := map[string]string{
		"DB_MAX_OPEN_CONNS":    "20",
		"DB_MAX_IDLE_CONNS":    "10",
		"DB_CONN_MAX_LIFETIME": "45m",
	}
	getenv := func(key string) string {
		return env[key]
	}
	got, err := loadDBPoolConfig(getenv)

	require.NoError(t, err)
	require.Equal(t, 20, got.maxOpenConns)
	require.Equal(t, 10, got.maxIdleConns)
	require.Equal(t, 45*time.Minute, got.connMaxLifetime)
}

func TestLoadDBPoolConfig_InvalidMaxOpen(t *testing.T) {
	env := map[string]string{
		"DB_MAX_OPEN_CONNS": "invalid",
	}
	getenv := func(key string) string {
		return env[key]
	}
	_, err := loadDBPoolConfig(getenv)

	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_MAX_OPEN_CONNS")
}

func TestLoadDBPoolConfig_InvalidMaxIdle(t *testing.T) {
	env := map[string]string{
		"DB_MAX_IDLE_CONNS": "invalid",
	}
	getenv := func(key string) string {
		return env[key]
	}
	_, err := loadDBPoolConfig(getenv)

	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_MAX_IDLE_CONNS")
}

func TestLoadDBPoolConfig_InvalidMaxLifetime(t *testing.T) {
	env := map[string]string{
		"DB_CONN_MAX_LIFETIME": "invalid",
	}
	getenv := func(key string) string {
		return env[key]
	}

	_, err := loadDBPoolConfig(getenv)

	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_CONN_MAX_LIFETIME")
}
