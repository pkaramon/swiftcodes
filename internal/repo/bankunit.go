package repo

import (
	"context"

	"github.com/pkarmon/swiftcodes/internal/model"
)

type BankUnit interface {
	Create(ctx context.Context, bank *model.BankUnit) error
	BulkCreate(ctx context.Context, banks []*model.BankUnit) error
	GetBySwiftCode(ctx context.Context, swiftCode model.SwiftCode) (*model.BankUnit, error)
	GetAllByCountry(ctx context.Context, countryISO2 model.CountryISO2) ([]*model.BankUnit, error)
	DeleteAll(ctx context.Context) error
	Delete(ctx context.Context, swiftCode model.SwiftCode) error
	GetAll(ctx context.Context) ([]*model.BankUnit, error)
	GetBranches(ctx context.Context, swiftCode model.SwiftCode) ([]*model.BankUnit, error)
}
