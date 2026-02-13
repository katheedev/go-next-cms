package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	return pgxpool.NewWithConfig(ctx, cfg)
}
