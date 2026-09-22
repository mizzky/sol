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
