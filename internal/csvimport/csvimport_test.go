package csvimport_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pkarmon/swiftcodes/internal/csvimport"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func configureTestDB(ctx context.Context, t *testing.T) postgres.DB {
	dbName, dbUser, dbPassword := "test", "test", "test"

	container, err := pgcontainer.Run(ctx,
		"postgres:latest",
		pgcontainer.WithDatabase(dbName),
		pgcontainer.WithUsername(dbUser),
		pgcontainer.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(container)
		assert.NoError(t, err)
	})

	assert.NoError(t, err)

	host, err := container.Host(ctx)
	assert.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port.Port(), dbName,
	)
	db, err := postgres.Connect(connStr)
	assert.NoError(t, err)
	err = db.SetupSchema(ctx)
	require.NoError(t, err)

	return db
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

func TestImportCountries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := configureTestDB(ctx, t)
	countryRepo := postgres.NewCountryRepo(db)

	t.Run("invalid csv data", func(t *testing.T) {
		csvData := `Name,Code
Poland,PLL`

		err := csvimport.Countries(ctx, strings.NewReader(csvData), countryRepo)

		assert.Error(t, err)
		countries, err := countryRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, countries, 0)
	})

	t.Run("valid csv data", func(t *testing.T) {
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
	t.Parallel()

	ctx := context.Background()
	db := configureTestDB(ctx, t)
	bankUnitRepo := postgres.NewBankUnitRepo(db)

	// we need to import countries first
	countryRepo := postgres.NewCountryRepo(db)
	err := csvimport.Countries(ctx, strings.NewReader(`Name,Code
Poland,PL`), countryRepo)
	assert.NoError(t, err)

	t.Run("invalid csv data", func(t *testing.T) {
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
