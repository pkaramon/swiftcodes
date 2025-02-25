package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pkarmon/swiftcodes/internal/database"
	"github.com/pkarmon/swiftcodes/internal/model"
)

type CountryRepo struct {
	db database.DB
}

func NewCountryRepo(db database.DB) *CountryRepo {
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
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM countries WHERE code = $1 AND name = $2)",
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

		code, err := model.NewCountryISO2(iso2)
		if err != nil {
			return nil, fmt.Errorf("invalid data present in db: %w", err)
		}

		country, err := model.NewCountry(code, name)
		if err != nil {
			return nil, fmt.Errorf("invalid data present in db: %w", err)
		}

		countries = append(countries, country)
	}

	return countries, nil
}
