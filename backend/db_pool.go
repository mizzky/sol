package main

import (
	"strconv"
	"time"
)

type dbPoolConfig struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

func loadDBPoolConfig(getenv func(string) string) (dbPoolConfig, error) {
	config := dbPoolConfig{
		maxOpenConns:    10,
		maxIdleConns:    5,
		connMaxLifetime: 30 * time.Minute,
	}

	if value := getenv("DB_MAX_OPEN_CONNS"); value != "" {
		config.maxOpenConns, _ = strconv.Atoi(value)
	}

	if value := getenv("DB_MAX_IDLE_CONNS"); value != "" {
		config.maxIdleConns, _ = strconv.Atoi(value)
	}

	if value := getenv("DB_CONN_MAX_LIFETIME"); value != "" {
		config.connMaxLifetime, _ = time.ParseDuration(value)
	}
	return config, nil
}
