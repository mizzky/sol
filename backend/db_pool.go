package main

import "time"

type dbPoolConfig struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

func loadDBPoolConfig(getenv func(string) string) (dbPoolConfig, error) {
	return dbPoolConfig{
		maxOpenConns:    10,
		maxIdleConns:    5,
		connMaxLifetime: 30 * time.Minute,
	}, nil
}
