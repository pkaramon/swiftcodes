package csvimport_test

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/pkarmon/swiftcodes/internal/csvimport"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/postgres"
	"github.com/pkarmon/swiftcodes/internal/repo"
	"github.com/stretchr/testify/assert"
)

var (
	db           postgres.DB
	bankUnitRepo repo.BankUnit
	countryRepo  repo.Country
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	createdDb, cleanup, err := postgres.ConfigureTestDB(ctx)
	if err != nil {
		log.Fatalf("could not seutp test db: %s", err)
	}
	defer cleanup()

	db = createdDb
	bankUnitRepo = postgres.NewBankUnitRepo(db)
	countryRepo = postgres.NewCountryRepo(db)

	m.Run()
}

func TestImportCountries(t *testing.T) {
	ctx := context.Background()
	clearDB := func(t *testing.T) func() {
		return func() {
			err := db.DropSchema(ctx)
			assert.NoError(t, err)
			err = db.SetupSchema(ctx)
			assert.NoError(t, err)
		}
	}

	t.Run("invalid csv data", func(t *testing.T) {
		t.Cleanup(clearDB(t))

		csvData := `Name,Code
Poland,PLL`

		err := csvimport.Countries(ctx, strings.NewReader(csvData), countryRepo)

		assert.Error(t, err)
		countries, err := countryRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, countries, 0)
	})

	t.Run("valid csv data", func(t *testing.T) {
		t.Cleanup(clearDB(t))

		csvData := `Name,Code
Poland,PL
Germany,DE`

		err := csvimport.Countries(ctx, strings.NewReader(csvData), countryRepo)

		assert.NoError(t, err)
		countries, err := countryRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, countries, 2)

		assert.NoError(t, err)
		pl := mustNewCountry(t, "PL", "Poland")
		de := mustNewCountry(t, "DE", "Germany")

		assert.Contains(t, countries, pl)
		assert.Contains(t, countries, de)

	})

}

func TestImportBankUnits(t *testing.T) {
	ctx := context.Background()
	err := csvimport.Countries(ctx, strings.NewReader(`Name,Code
Poland,PL`), countryRepo)
	assert.NoError(t, err)

	clearDB := func(t *testing.T) func() {
		return func() {
			assert.NoError(t, bankUnitRepo.DeleteAll(ctx))
		}
	}

	t.Run("invalid csv data", func(t *testing.T) {
		t.Cleanup(clearDB(t))
		// it's invalid because of PLLL, iso2 code must be 2 characters long
		csvData := `COUNTRY ISO2 CODE,SWIFT CODE,CODE TYPE,NAME,ADDRESS,TOWN NAME,COUNTRY NAME,TIME ZONE
PLLL,BIGBPLPWCUS,BIC11,BANK MILLENNIUM S.A.,"HARMONY CENTER UL. STANISLAWA ZARYNA 2A WARSZAWA, MAZOWIECKIE, 02-593",WARSZAWA,POLAND,Europe/Warsaw
PL,HYVEPLP2XXX,BIC11,PEKAO BANK HIPOTECZNY SA,"RENAISSANCE TOWER UL. SKIERNIEWICKA 10A WARSZAWA, MAZOWIECKIE, 01-230",WARSZAWA,POLAND,Europe/Warsaw`

		err := csvimport.BankUnits(ctx, strings.NewReader(csvData), bankUnitRepo)

		assert.Error(t, err)
		bankUnits, err := bankUnitRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, bankUnits, 0)
	})

	t.Run("valid csv data", func(t *testing.T) {
		t.Cleanup(clearDB(t))

		csvData := `COUNTRY ISO2 CODE,SWIFT CODE,CODE TYPE,NAME,ADDRESS,TOWN NAME,COUNTRY NAME,TIME ZONE
PL,BIGBPLPWCUS,BIC11,BANK MILLENNIUM S.A.,"HARMONY CENTER UL. STANISLAWA ZARYNA 2A WARSZAWA, MAZOWIECKIE, 02-593",WARSZAWA,POLAND,Europe/Warsaw
PL,HYVEPLP2XXX,BIC11,PEKAO BANK HIPOTECZNY SA,"RENAISSANCE TOWER UL. SKIERNIEWICKA 10A WARSZAWA, MAZOWIECKIE, 01-230",WARSZAWA,POLAND,Europe/Warsaw`

		err := csvimport.BankUnits(ctx, strings.NewReader(csvData), bankUnitRepo)

		assert.NoError(t, err)
		bankUnits, err := bankUnitRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, bankUnits, 2)

		milienium := mustNewBankUnit(t, "BIGBPLPWCUS", "PL", "POLAND", "BANK MILLENNIUM S.A.", "HARMONY CENTER UL. STANISLAWA ZARYNA 2A WARSZAWA, MAZOWIECKIE, 02-593", false)
		pekao := mustNewBankUnit(t, "HYVEPLP2XXX", "PL", "POLAND", "PEKAO BANK HIPOTECZNY SA", "RENAISSANCE TOWER UL. SKIERNIEWICKA 10A WARSZAWA, MAZOWIECKIE, 01-230", true)

		bankUnitsValues := []model.BankUnit{*bankUnits[0], *bankUnits[1]}

		assert.Contains(t, bankUnitsValues, *milienium)
		assert.Contains(t, bankUnitsValues, *pekao)
	})

}

func mustNewCountry(t *testing.T, code, name string) model.Country {
	country, err := model.NewCountry(code, name)
	assert.NoError(t, err)
	return country
}

func mustNewBankUnit(t *testing.T, swiftCode, countryISO2, countryName, bankName, address string, isHeadquarter bool) *model.BankUnit {
	bankUnit, err := model.NewBankUnit(swiftCode, countryISO2, countryName, address, bankName, isHeadquarter)
	assert.NoError(t, err)
	return bankUnit
}
