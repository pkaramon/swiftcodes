package repo

import (
	"context"

	"github.com/pkarmon/swiftcodes/internal/model"
)

type BankUnit interface {
	Create(ctx context.Context, bank *model.BankUnit) error
	BulkCreate(ctx context.Context, banks []*model.BankUnit) error
	GetBySwiftCode(ctx context.Context, swiftCode model.SwiftCode) (*model.BankUnit, error)
	ListByCountry(ctx context.Context, countryISO2 model.CountryISO2) ([]*model.BankUnit, error)
	Delete(ctx context.Context, swiftCode model.SwiftCode) error
	GetAll(ctx context.Context) ([]*model.BankUnit, error)
}
