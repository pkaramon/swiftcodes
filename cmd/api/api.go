package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pkarmon/swiftcodes/internal/csvimport"
	"github.com/pkarmon/swiftcodes/internal/database"
	"github.com/pkarmon/swiftcodes/internal/handlers"
	"github.com/pkarmon/swiftcodes/internal/postgres"
)

func main() {
	db, err := database.Connect("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.DropSchema(ctx); err != nil {
		log.Fatal(err)
	}

	if err := db.SetupSchema(ctx); err != nil {
		log.Fatal(err)
	}

	bankRepo := postgres.NewBankUnitRepo(db)
	countryRepo := postgres.NewCountryRepo(db)

	countrycodes, err := os.Open("initialData/countries_iso3166b.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer countrycodes.Close()

	if err := csvimport.Countries(ctx, countrycodes, countryRepo); err != nil {
		log.Fatal(err)
	}

	bankunits, err := os.Open("initialData/swiftcodes.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer bankunits.Close()

	if err := csvimport.BankUnits(ctx, bankunits, bankRepo); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/v1/swift-codes/").Subrouter()

	api.HandleFunc("/{country}/{countryISO2code}",
		handlers.GetAllBankUnitsForCountry(bankRepo)).Methods(http.MethodGet)
	api.HandleFunc("/{swiftCode}",
		handlers.GetBankUnit(bankRepo)).Methods(http.MethodGet)
	api.HandleFunc("/", handlers.CreateBankUnit(bankRepo, countryRepo)).Methods(http.MethodPost)
	api.HandleFunc("/{swiftCode}", handlers.DeleteBankUnit(bankRepo)).Methods(http.MethodDelete)

	http.ListenAndServe(":8080", r)
}
