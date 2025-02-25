package model_test

import (
	"testing"

	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCountry_ValidatesEmptyName(t *testing.T) {
	t.Parallel()

	code, err := model.NewCountryISO2("PL")
	assert.NoError(t, err)
	country, err := model.NewCountry(code, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "country name cannot be empty")
	assert.Nil(t, country)
}

func TestCountry_ConvertsNameToUppercase(t *testing.T) {
	t.Parallel()

	code, err := model.NewCountryISO2("PL")
	assert.NoError(t, err)
	country, err := model.NewCountry(code, "poland")

	assert.NoError(t, err)
	assert.NotNil(t, country)
	assert.Equal(t, "POLAND", country.Name)
}
