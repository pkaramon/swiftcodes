package model_test

import (
	"testing"

	"github.com/pkarmon/swiftcodes/internal/model"
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
			swiftCode:     "BPKOPLPWXXX",
			countryISO2:   "PL",
			countryName:   "POLAND",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
		},
		{
			name:          "empty country name",
			swiftCode:     "BPKOPLPWXXX",
			countryISO2:   "PL",
			countryName:   "",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
			wantErr:       "country name cannot be empty",
		},
		{
			name:          "invalid country iso2 code",
			swiftCode:     "BPKOPLPWXXX",
			countryISO2:   "PL3",
			countryName:   "POLAND",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
			wantErr:       "country ISO2 code length must be 2 characters",
		},
		{
			name:          "valid branch",
			swiftCode:     "BREXPLPW022",
			countryISO2:   "PL",
			countryName:   "POLAND",
			address:       "Krakow",
			bankName:      "mBank Krakow",
			isHeadquarter: false,
		},
		{
			name:        "invalid swift code length",
			swiftCode:   "BPKO",
			countryISO2: "PL",
			countryName: "POLAND",
			address:     "Warsaw",
			bankName:    "PKO Bank Polski",
			wantErr:     "swift code length must be 11 characters",
		},
		{
			name:        "country mismatch",
			swiftCode:   "BPKODEPWXXX",
			countryISO2: "PL",
			countryName: "POLAND",
			address:     "Warsaw",
			bankName:    "PKO Bank Polski",
			wantErr:     "swift code and country ISO2 code mismatch",
		},
		{
			name:          "invalid headquarter code",
			swiftCode:     "BPKOPLPW022",
			countryISO2:   "PL",
			countryName:   "POLAND",
			address:       "Warsaw",
			bankName:      "PKO Bank Polski",
			isHeadquarter: true,
			wantErr:       "headquarter must have branch code XXX",
		},
		{
			name:        "empty bank name",
			swiftCode:   "BPKOPLPWXXX",
			countryISO2: "PL",
			countryName: "POLAND",
			address:     "Warsaw",
			bankName:    "",
			wantErr:     "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := model.NewBankUnit(tt.swiftCode, tt.countryISO2, tt.countryName, tt.address, tt.bankName, tt.isHeadquarter)

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
			assert.Equal(t, tt.countryISO2, got.Country.Code.String())
			assert.Equal(t, tt.address, got.Address)
		})
	}
}

func TestSwiftCodeHasHeadquartersBranchCode(t *testing.T) {
	hq, err := model.NewSwiftCode("BPKOPLPWXXX")
	assert.NoError(t, err)
	assert.True(t, hq.HasHeadQuartersBranchCode())

	branch, err := model.NewSwiftCode("BPKOPLPW022")
	assert.NoError(t, err)
	assert.False(t, branch.HasHeadQuartersBranchCode())
}

func TestSwiftCodeBaseCodeAndBranchCode(t *testing.T) {
	sc, err := model.NewSwiftCode("BPKOPLPWXXX")
	assert.NoError(t, err)
	assert.Equal(t, "BPKOPLPW", sc.BaseCode())
	assert.Equal(t, "XXX", sc.BranchCode())
}

func TestSwiftCodeString(t *testing.T) {
	sc, err := model.NewSwiftCode("BPKOPLPWXXX")
	assert.NoError(t, err)
	assert.Equal(t, "BPKOPLPWXXX", sc.String())
}
