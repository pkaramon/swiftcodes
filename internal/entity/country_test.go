package entity_test

import (
	"testing"

	"github.com/pkarmon/swiftcodes/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestCountry_ValidatesEmptyName(t *testing.T) {
	t.Parallel()

	code, err := entity.NewCountryISO2("PL")
	assert.NoError(t, err)
	country, err := entity.NewCountry(code, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "country name cannot be empty")
	assert.Nil(t, country)
}

func TestCountry_ConvertsNameToUppercase(t *testing.T) {
	t.Parallel()

	code, err := entity.NewCountryISO2("PL")
	assert.NoError(t, err)
	country, err := entity.NewCountry(code, "poland")

	assert.NoError(t, err)
	assert.NotNil(t, country)
	assert.Equal(t, "POLAND", country.Name)
}
