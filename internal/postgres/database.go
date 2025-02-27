package postgres

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

func (db *DB) Ping(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}
	return nil
}

func (db *DB) SetupSchema(ctx context.Context) error {
	return db.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, setupDatabase)
		return err
	})
}

func (db *DB) DropSchema(ctx context.Context) error {
	return db.InTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DROP VIEW IF EXISTS bank_units_with_country;")
		if err != nil {
			return fmt.Errorf("failed to drop view: %w", err)
		}
		_, err = tx.Exec(ctx, "DROP TABLE IF EXISTS bank_units; DROP TABLE IF EXISTS countries;")
		if err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
		return err
	})
}

func (db *DB) RestartSchema(ctx context.Context) error {
	if err := db.DropSchema(ctx); err != nil {
		return err
	}
	return db.SetupSchema(ctx)
}

func (db *DB) InTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
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
);

CREATE TABLE IF NOT EXISTS bank_units (
    id SERIAL PRIMARY KEY,
    country_iso2 CHAR(2) NOT NULL REFERENCES countries(iso2),
    swift_code CHAR(11) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
	address TEXT NOT NULL,
    is_headquarter BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bank_units_country_iso2 ON bank_units (country_iso2);
CREATE INDEX IF NOT EXISTS idx_bank_units_swift_code ON bank_units (swift_code);
CREATE INDEX IF NOT EXISTS idx_bank_units_base_code ON bank_units (LEFT(swift_code, 8));

CREATE OR REPLACE VIEW bank_units_with_country AS
SELECT
    bu.id,
    bu.country_iso2,
    bu.swift_code,
    bu.name as bank_name,
    bu.address,
    bu.is_headquarter,
    c.name as country_name
FROM bank_units bu
JOIN countries c ON bu.country_iso2 = c.iso2;
`
