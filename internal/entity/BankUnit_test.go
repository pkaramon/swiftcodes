package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBankUnit(t *testing.T) {
	tests := []struct {
		name          string
		swiftCode     string
		countryISO2   string
		countryName   string
		address       string
		bankName      string
		isHeadquarter bool
		wantErr       string
	}{
		{
			name:          "valid headquarter",
			swiftCode:     "DEUTDEFFXXX",
			countryISO2:   "DE",
			countryName:   "GERMANY",
			address:       "Frankfurt",
			bankName:      "Deutsche Bank",
			isHeadquarter: true,
		},
		{
			name:          "valid branch",
			swiftCode:     "DEUTDEFF500",
			countryISO2:   "DE",
			countryName:   "GERMANY",
			address:       "Berlin",
			bankName:      "Deutsche Bank Berlin",
			isHeadquarter: false,
		},
		{
			name:        "invalid swift code length",
			swiftCode:   "DEUT",
			countryISO2: "DE",
			countryName: "GERMANY",
			address:     "Berlin",
			bankName:    "Deutsche Bank",
			wantErr:     "swift code length must be 11 characters",
		},
		{
			name:        "country mismatch",
			swiftCode:   "DEUTFRFFXXX",
			countryISO2: "DE",
			countryName: "GERMANY",
			address:     "Berlin",
			bankName:    "Deutsche Bank",
			wantErr:     "swift code and country ISO2 code mismatch",
		},
		{
			name:          "invalid headquarter code",
			swiftCode:     "DEUTDEFF500",
			countryISO2:   "DE",
			countryName:   "GERMANY",
			address:       "Berlin",
			bankName:      "Deutsche Bank",
			isHeadquarter: true,
			wantErr:       "headquarter must have branch code XXX",
		},
		{
			name:        "empty bank name",
			swiftCode:   "DEUTDEFFXXX",
			countryISO2: "DE",
			countryName: "GERMANY",
			address:     "Berlin",
			bankName:    "",
			wantErr:     "name cannot be empty",
		},
		{
			name:        "empty country name",
			swiftCode:   "DEUTDEFFXXX",
			countryISO2: "DE",
			countryName: "",
			address:     "Berlin",
			bankName:    "Deutsche Bank",
			wantErr:     "country name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewBankUnit(tt.swiftCode, tt.countryISO2, tt.countryName, tt.address, tt.bankName, tt.isHeadquarter)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.isHeadquarter, got.IsHeadquarter)
			assert.Equal(t, tt.bankName, got.Name)
			assert.Equal(t, tt.countryName, got.Country.Name)
			assert.Equal(t, tt.countryISO2, got.Country.CodeISO2)
			assert.Equal(t, tt.address, got.Address)
		})
	}
}
