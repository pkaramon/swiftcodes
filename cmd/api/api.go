package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/pkarmon/swiftcodes/internal/csvimport"
	"github.com/pkarmon/swiftcodes/internal/handlers"
	"github.com/pkarmon/swiftcodes/internal/middleware"
	"github.com/pkarmon/swiftcodes/internal/postgres"
)

func main() {
	// Load configurations
	serverCfg := LoadServerConfig()
	dbConnStr := LoadDatabaseConnectionStr()

	// Connect to database
	db, err := postgres.Connect(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Ping database
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	// Setup database schema
	if err := db.DropSchema(ctx); err != nil {
		log.Fatal(err)
	}
	if err := db.SetupSchema(ctx); err != nil {
		log.Fatal(err)
	}

	// Initialize repos and import data
	if err := setupInitialData(ctx, db); err != nil {
		log.Fatal(err)
	}

	// Configure and start server
	srv := setupServer(serverCfg, db)
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), serverCfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func setupServer(cfg ServerConfig, db postgres.DB) *http.Server {
	bankRepo := postgres.NewBankUnitRepo(db)
	countryRepo := postgres.NewCountryRepo(db)

	r := mux.NewRouter()
	api := r.PathPrefix("/v1/swift-codes").Subrouter()

	api.HandleFunc("/country/{countryISO2code}",
		handlers.GetAllBankUnitsForCountry(bankRepo, countryRepo)).Methods(http.MethodGet)
	api.HandleFunc("/{swiftCode}",
		handlers.GetBankUnit(bankRepo)).Methods(http.MethodGet)
	api.HandleFunc("/{swiftCode}",
		handlers.DeleteBankUnit(bankRepo)).Methods(http.MethodDelete)
	api.HandleFunc("/",
		handlers.CreateBankUnit(bankRepo, countryRepo)).Methods(http.MethodPost)

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      middleware.Logging(r),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}

func setupInitialData(ctx context.Context, db postgres.DB) error {
	bankRepo := postgres.NewBankUnitRepo(db)
	countryRepo := postgres.NewCountryRepo(db)

	countrycodes, err := os.Open("initialData/countries_iso3166b.csv")
	if err != nil {
		return fmt.Errorf("open countries file: %w", err)
	}
	defer countrycodes.Close()

	if err := csvimport.Countries(ctx, countrycodes, countryRepo); err != nil {
		return fmt.Errorf("import countries: %w", err)
	}

	bankunits, err := os.Open("initialData/swiftcodes.csv")
	if err != nil {
		return fmt.Errorf("open bank units file: %w", err)
	}
	defer bankunits.Close()

	if err := csvimport.BankUnits(ctx, bankunits, bankRepo); err != nil {
		return fmt.Errorf("import bank units: %w", err)
	}

	return nil
}
