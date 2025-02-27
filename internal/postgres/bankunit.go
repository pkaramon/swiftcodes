package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/repo"
)

type BankUnitRepo struct {
	db DB
}

func NewBankUnitRepo(db DB) *BankUnitRepo {
	return &BankUnitRepo{db: db}
}

type bankUnitRecord struct {
	ID            int    `db:"id"`
	CountryISO2   string `db:"country_iso2"`
	CountryName   string `db:"country_name"`
	SwiftCode     string `db:"swift_code"`
	Name          string `db:"bank_name"`
	Address       string `db:"address"`
	IsHeadquarter bool   `db:"is_headquarter"`
}

func (rec *bankUnitRecord) toModel() (*model.BankUnit, error) {
	return model.NewBankUnit(
		rec.SwiftCode,
		rec.CountryISO2,
		rec.CountryName,
		rec.Address,
		rec.Name,
		rec.IsHeadquarter,
	)
}

func (r *BankUnitRepo) BulkCreate(ctx context.Context, bankUnits []*model.BankUnit) error {
	return r.db.InTx(ctx, func(tx pgx.Tx) error {
		rows := make([][]interface{}, len(bankUnits))
		for i, bankUnit := range bankUnits {
			rows[i] = []interface{}{bankUnit.Country.Code.String(), bankUnit.SwiftCode.String(), bankUnit.Name, bankUnit.Address, bankUnit.IsHeadquarter}
		}

		_, err := tx.CopyFrom(ctx, pgx.Identifier{"bank_units"},
			[]string{"country_iso2", "swift_code", "name", "address", "is_headquarter"}, pgx.CopyFromRows(rows))
		if err != nil {
			return fmt.Errorf("failed to copy bank units: %w", err)
		}
		return nil
	})
}

func (r *BankUnitRepo) Create(ctx context.Context, bankUnit *model.BankUnit) error {
	_, err := r.db.Exec(ctx, `
		INSERT 
		INTO bank_units
		(country_iso2, swift_code, name, address, is_headquarter)
		VALUES ($1, $2, $3, $4, $5)`,
		bankUnit.Country.Code.String(), bankUnit.SwiftCode.String(), bankUnit.Name, bankUnit.Address, bankUnit.IsHeadquarter)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return repo.ErrDuplicate
		}
		return fmt.Errorf("failed to create bank unit: %w", err)
	}
	return nil
}

func (r *BankUnitRepo) Delete(ctx context.Context, swiftCode model.SwiftCode) error {
	_, err := r.db.Exec(ctx, "DELETE FROM bank_units WHERE swift_code = $1", swiftCode.String())
	if err != nil {
		return fmt.Errorf("failed to delete bank unit: %w", err)
	}
	return nil
}

func (r *BankUnitRepo) GetBySwiftCode(ctx context.Context, swiftCode model.SwiftCode) (*model.BankUnit, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM bank_units_with_country
		WHERE swift_code = $1`,
		swiftCode.String())

	if err != nil {
		return nil, fmt.Errorf("failed to get bank unit: %w", err)
	}

	rec, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[bankUnitRecord])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to collect bank unit: %w", err)
	}

	unit, err := rec.toModel()
	if err != nil {
		return nil, err
	}

	return unit, nil
}

func (r *BankUnitRepo) GetAllByCountry(ctx context.Context, countryISO2 model.CountryISO2) ([]*model.BankUnit, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM bank_units_with_country
		WHERE country_iso2 = $1
		`,
		countryISO2.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list bank units: %w", err)
	}
	return r.fromRowsToModels(rows)
}

func (r *BankUnitRepo) GetAll(ctx context.Context) ([]*model.BankUnit, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM bank_units_with_country`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all bank units: %w", err)
	}

	return r.fromRowsToModels(rows)
}

func (r *BankUnitRepo) GetBranches(ctx context.Context, swiftCode model.SwiftCode) ([]*model.BankUnit, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM bank_units_with_country
		WHERE LEFT(swift_code, 8) = $1 AND swift_code != $2
	`, swiftCode.BaseCode(), swiftCode.String())

	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	return r.fromRowsToModels(rows)
}

func (r *BankUnitRepo) fromRowsToModels(rows pgx.Rows) ([]*model.BankUnit, error) {
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[bankUnitRecord])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, repo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to collect bank units: %w", err)
	}

	result := make([]*model.BankUnit, len(records))
	for i, rec := range records {
		unit, err := rec.toModel()
		if err != nil {
			return nil, fmt.Errorf("failed to map bank unit record: %w", err)
		}
		result[i] = unit
	}

	return result, nil
}
