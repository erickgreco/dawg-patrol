package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	MaxConns              int32
	MinConns              int32
	MaxConnLifeTime       time.Duration
	MaxConnIdleTime       time.Duration
	HealthCheckPeriod     time.Duration
	MaxConnLifeTimeJitter time.Duration
}

//! iMPORTANT: IMPLEMENT DEFAULT CONFIG TO AVOID EMPTY STRUCTS BREAKING DB

// Stablishes database connection and config
func DBConn(connString string, cfg DBConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifeTime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod
	config.MaxConnLifetimeJitter = cfg.MaxConnLifeTimeJitter

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
