package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/pkarmon/swiftcodes/internal/csvmapper"
	"github.com/pkarmon/swiftcodes/internal/database"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/postgres"
	"github.com/pkarmon/swiftcodes/internal/repo"
)

func main() {
	f, err := os.Open("initialData/countries_iso3166b.csv")
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.SetupSchema(context.Background()); err != nil {
		log.Fatal(err)
	}

	countryRepo := postgres.NewCountryRepo(db)
	if err := importCountries(context.Background(), f, countryRepo); err != nil {
		log.Fatal(err)
	}

	ctrs, err := countryRepo.GetAll(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range ctrs {
		log.Printf("Country: %s", c)
	}

}

func importCountries(ctx context.Context, src io.Reader, r repo.Country) error {
	mapper := csvmapper.New(src, []string{"Name", "Code"}, func(record []string) (model.Country, error) {
		var country model.Country
		iso2, err := model.NewCountryISO2(record[1])
		if err != nil {
			return country, err
		}

		country, err = model.NewCountry(iso2, record[0])
		if err != nil {
			return country, err
		}
		return country, nil
	})

	countries, err := mapper.MapAll()
	if err != nil {
		return err
	}

	log.Printf("Sucessfully loaded %d countries", len(countries))

	if err := r.BulkCreate(ctx, countries); err != nil {
		return err
	}

	log.Println("Countries imported successfully")

	return nil
}
