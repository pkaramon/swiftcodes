package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pkarmon/swiftcodes/internal/database"
	"github.com/pkarmon/swiftcodes/internal/entity"
)

type CountryRepository struct {
	db *database.DB
}

func NewCountryRepository(db *database.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

func (r *CountryRepository) BulkCreate(ctx context.Context, countries []entity.Country) error {
	return r.db.InTx(ctx, func(tx pgx.Tx) error {
		rows := make([][]interface{}, len(countries))
		for i, country := range countries {
			rows[i] = []interface{}{country.Code.String(), country.Name}
		}

		_, err := tx.CopyFrom(ctx, pgx.Identifier{"countries"}, []string{"code", "name"}, pgx.CopyFromRows(rows))
		if err != nil {
			return fmt.Errorf("failed to copy countries: %w", err)
		}
		return nil
	})
}

func (r *CountryRepository) Exists(ctx context.Context, country entity.Country) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM countries WHERE code = $1 AND name = $2)",
		country.Code.String(),
		country.Name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if country exists: %w", err)
	}

	return exists, nil

}
