package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pkarmon/swiftcodes/internal/database"
	"github.com/pkarmon/swiftcodes/internal/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func configureTestDB(ctx context.Context, t *testing.T) database.DB {
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
	db, err := database.Connect(connStr)
	assert.NoError(t, err)
	err = db.SetupSchema(ctx)
	assert.NoError(t, err)

	return db
}

func TestImportCountries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := configureTestDB(ctx, t)
	countryRepo := postgres.NewCountryRepo(db)

	csvData := `Name,Code
Poland,PL
Germany,DE`

	reader := strings.NewReader(csvData)
	err := importCountries(ctx, reader, countryRepo)
	assert.NoError(t, err)

	countries, err := countryRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, countries, 2)

	expected := map[string]string{
		"PL": "POLAND",
		"DE": "GERMANY",
	}

	for _, c := range countries {
		assert.Equal(t, expected[c.Code.String()], c.Name)
	}
}
