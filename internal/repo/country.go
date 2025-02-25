package repo

import (
	"context"

	"github.com/pkarmon/swiftcodes/internal/model"
)

type Country interface {
	BulkCreate(ctx context.Context, countries []model.Country) error
	Exists(ctx context.Context, country model.Country) (bool, error)
	GetAll(ctx context.Context) ([]model.Country, error)
}
