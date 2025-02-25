package repository

import (
	"context"

	"github.com/pkarmon/swiftcodes/internal/entity"
)

type BankUnitRepository interface {
	Create(ctx context.Context, bank *entity.BankUnit) error
	BulkCreate(ctx context.Context, banks []*entity.BankUnit) error
	GetBySwiftCode(ctx context.Context, swiftCode entity.SwiftCode) (*entity.BankUnit, error)
	ListByCountry(ctx context.Context, countryISO2 entity.CountryISO2) ([]*entity.BankUnit, error)
	Update(ctx context.Context, bank *entity.BankUnit) error
	Delete(ctx context.Context, swiftCode entity.SwiftCode) error
}
