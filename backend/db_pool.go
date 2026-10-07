package main

import (
	"fmt"
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
		maxOpenConns, err := strconv.Atoi(value)
		if err != nil {
			return dbPoolConfig{}, fmt.Errorf(
				"DB_MAX_OPEN_CONNS must be an integer: %w",
				err,
			)
		}
		config.maxOpenConns = maxOpenConns
	}

	if value := getenv("DB_MAX_IDLE_CONNS"); value != "" {
		maxIdleConns, err := strconv.Atoi(value)
		if err != nil {
			return dbPoolConfig{}, fmt.Errorf(
				"DB_MAX_IDLE_CONNS must be an integer: %w",
				err,
			)
		}
		config.maxIdleConns = maxIdleConns
	}

	if value := getenv("DB_CONN_MAX_LIFETIME"); value != "" {
		connMaxLifetime, err := time.ParseDuration(value)
		if err != nil {
			return dbPoolConfig{}, fmt.Errorf(
				"DB_CONN_MAX_LIFETIME must be a duration: %w",
				err,
			)
		}
		config.connMaxLifetime = connMaxLifetime
	}
	return config, nil
}
