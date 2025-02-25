package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}

func Connect(connectionURL string) (DB, error) {
	cfg, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return DB{}, fmt.Errorf("failed to parse connection string: %w", err)
	}

	cfg.ConnConfig.ConnectTimeout = 10 * time.Second
	cfg.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return DB{}, fmt.Errorf("failed to create Pool: %w ", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return DB{}, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return DB{pool}, nil
}

func (p *DB) SetupSchema(ctx context.Context) error {
	return p.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, setupDatabase)
		return err
	})
}

func (p *DB) InTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

const setupDatabase = `
CREATE TABLE IF NOT EXISTS countries (
	iso2 CHAR(2) PRIMARY KEY,
	name VARCHAR(255) NOT NULL
)

CREATE TABLE IF NOT EXISTS bank_units (
    id SERIAL PRIMARY KEY,
    country_iso2 CHAR(2) NOT NULL REFERENCES countries(iso2),
    swift_code CHAR(11) NOT NULL,
    name VARCHAR(255) NOT NULL,
	address TEXT NOT NULL,
    is_headquarter BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_swift_codes_country_iso2 ON swift_codes (country_iso2);
CREATE INDEX IF NOT EXISTS idx_swift_codes_swift_code ON swift_codes (swift_code);
CREATE INDEX IF NOT EXISTS idx_swift_codes_base_code ON swift_codes (LEFT(swift_code, 8));
`
