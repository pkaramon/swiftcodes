package entity_test

import (
	"testing"

	"github.com/pkarmon/swiftcodes/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewBankUnit(t *testing.T) {
	tests := []struct {
		name          string
		swiftCode     string
		countryISO2   string
		address       string
		bankName      string
		isHeadquarter bool
		wantErr       string
	}{
		{
			name:          "valid headquarter",
			swiftCode:     "BPKOPLPWXXX",
			countryISO2:   "PL",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
		},
		{
			name:          "valid branch",
			swiftCode:     "BREXPLPW022",
			countryISO2:   "PL",
			address:       "Krakow",
			bankName:      "mBank Krakow",
			isHeadquarter: false,
		},
		{
			name:        "invalid swift code length",
			swiftCode:   "BPKO",
			countryISO2: "PL",
			address:     "Warsaw",
			bankName:    "PKO Bank Polski",
			wantErr:     "swift code length must be 11 characters",
		},
		{
			name:        "country mismatch",
			swiftCode:   "BPKODEPWXXX",
			countryISO2: "PL",
			address:     "Warsaw",
			bankName:    "PKO Bank Polski",
			wantErr:     "swift code and country ISO2 code mismatch",
		},
		{
			name:          "invalid headquarter code",
			swiftCode:     "BPKOPLPW022",
			countryISO2:   "PL",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
			wantErr:       "headquarter must have branch code XXX",
		},
		{
			name:        "empty bank name",
			swiftCode:   "BPKOPLPWXXX",
			countryISO2: "PL",
			address:     "Warsaw",
			bankName:    "",
			wantErr:     "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := entity.NewBankUnit(tt.swiftCode, tt.countryISO2, tt.address, tt.bankName, tt.isHeadquarter)

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
			assert.Equal(t, tt.countryISO2, got.CountryISO2.String())
			assert.Equal(t, tt.address, got.Address)
		})
	}
}
