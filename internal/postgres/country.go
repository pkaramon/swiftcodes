package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/repo"
)

type CountryRepo struct {
	db DB
}

func NewCountryRepo(db DB) *CountryRepo {
	return &CountryRepo{db: db}
}

func (r *CountryRepo) BulkCreate(ctx context.Context, countries []model.Country) error {
	return r.db.InTx(ctx, func(tx pgx.Tx) error {
		rows := make([][]interface{}, len(countries))
		for i, country := range countries {
			rows[i] = []interface{}{country.Code.String(), country.Name}
		}

		_, err := tx.CopyFrom(ctx, pgx.Identifier{"countries"}, []string{"iso2", "name"}, pgx.CopyFromRows(rows))
		if err != nil {
			return fmt.Errorf("failed to copy countries: %w", err)
		}
		return nil
	})
}

func (r *CountryRepo) Exists(ctx context.Context, country model.Country) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM countries WHERE iso2 = $1 AND name = $2)",
		country.Code.String(),
		country.Name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if country exists: %w", err)
	}

	return exists, nil

}

func (r *CountryRepo) GetAll(ctx context.Context) ([]model.Country, error) {
	rows, err := r.db.Query(ctx, "SELECT iso2, name FROM countries")
	if err != nil {
		return nil, fmt.Errorf("failed to get countries: %w", err)
	}
	defer rows.Close()

	var iso2 string
	var name string

	countries := make([]model.Country, 0)
	for rows.Next() {
		err := rows.Scan(&iso2, &name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}

		country, err := model.NewCountry(iso2, name)
		if err != nil {
			return nil, fmt.Errorf("invalid data present in db: %w", err)
		}

		countries = append(countries, country)
	}

	return countries, nil
}

func (r *CountryRepo) GetByCode(ctx context.Context, code model.CountryISO2) (model.Country, error) {
	var iso2 string
	var name string

	err := r.db.QueryRow(ctx, "SELECT iso2, name FROM countries WHERE iso2 = $1", code.String()).Scan(&iso2, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Country{}, repo.ErrNotFound
	}
	if err != nil {
		return model.Country{}, fmt.Errorf("failed to get country by iso2: %w", err)
	}

	return model.NewCountry(iso2, name)
}
