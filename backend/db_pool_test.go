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

func TestLoadDBPoolConfig_RejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantEnv string
	}{
		{
			name:    "MaxOpenが0",
			env:     map[string]string{"DB_MAX_OPEN_CONNS": "0"},
			wantEnv: "DB_MAX_OPEN_CONNS",
		},
		{
			name:    "MaxOpenが負数",
			env:     map[string]string{"DB_MAX_OPEN_CONNS": "-1"},
			wantEnv: "DB_MAX_OPEN_CONNS",
		},
		{
			name:    "MaxIdleが負数",
			env:     map[string]string{"DB_MAX_IDLE_CONNS": "-1"},
			wantEnv: "DB_MAX_IDLE_CONNS",
		},
		{
			name:    "MaxLifetimeが負数",
			env:     map[string]string{"DB_CONN_MAX_LIFETIME": "-1s"},
			wantEnv: "DB_CONN_MAX_LIFETIME",
		},
		{
			name: "MaxIdleがMaxOpenを超える",
			env: map[string]string{
				"DB_MAX_OPEN_CONNS": "5",
				"DB_MAX_IDLE_CONNS": "6",
			},
			wantEnv: "DB_MAX_IDLE_CONNS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				return tt.env[key]
			}

			_, err := loadDBPoolConfig(getenv)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantEnv)

		})
	}
}

func TestLoadDBPoolConfig_AllowsZeroIdleAndLifetime(t *testing.T) {
	env := map[string]string{
		"DB_MAX_IDLE_CONNS":    "0",
		"DB_CONN_MAX_LIFETIME": "0s",
	}
	getenv := func(key string) string {
		return env[key]
	}

	got, err := loadDBPoolConfig(getenv)

	require.NoError(t, err)
	require.Equal(t, 0, got.maxIdleConns)
	require.Equal(t, time.Duration(0), got.connMaxLifetime)
}
