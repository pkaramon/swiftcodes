package repository

import (
	"context"

	"github.com/pkarmon/swiftcodes/internal/entity"
)

type CountryRepository interface {
	BulkCreate(ctx context.Context, countries []entity.Country) error
	Exists(ctx context.Context, country entity.Country) (bool, error)
}
