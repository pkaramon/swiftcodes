package handlers_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkarmon/swiftcodes/internal/handlers"
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
	exitIfErr(err)
	defer cleanup()

	db = createdDb
	bankUnitRepo = postgres.NewBankUnitRepo(db)
	countryRepo = postgres.NewCountryRepo(db)

	resetTestData(ctx)

	m.Run()
}

func resetTestData(ctx context.Context) {
	_ = db.DropSchema(context.Background())
	exitIfErr(db.SetupSchema(ctx))

	pl := Must(model.NewCountry("PL", "POLAND"))
	bg := Must(model.NewCountry("BG", "BULGARIA"))
	de := Must(model.NewCountry("DE", "GERMANY"))

	exitIfErr(countryRepo.BulkCreate(ctx, []model.Country{pl, bg, de}))

	hq := Must(model.NewBankUnit(
		"BPKOPLPWXXX",
		"PL",
		"POLAND",
		"UL. PULAWSKA 15  WARSZAWA, MAZOWIECKIE, 02-515",
		"PKO BANK POLSKI S.A.",
		true,
	))

	branch1 := Must(model.NewBankUnit(
		"BPKOPLPWCSD",
		"PL",
		"POLAND",
		"WARSZAWA, MAZOWIECKIE",
		"PKO BANK POLSKI S.A.",
		false,
	))

	branch2 := Must(model.NewBankUnit(
		"BPKOPLPWGDG",
		"PL",
		"POLAND",
		"SWIETOJANSKA 17  GDYNIA, POMORSKIE, 71-368",
		"PKO BANK POLSKI S.A.",
		false,
	))

	bulgarian := Must(model.NewBankUnit(
		"BEFNBGS1XXX",
		"BG",
		"BULGARIA",
		"VISKIAR PLANINA 19 FLOOR 2 SOFIA, SOFIA, 1407",
		"BENCHMARK FINANCE",
		true,
	))

	exitIfErr(bankUnitRepo.BulkCreate(ctx, []*model.BankUnit{hq, branch1, branch2, bulgarian}))
}

func TestGetBankUnit(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{swiftCode}", handlers.GetBankUnit(bankUnitRepo)).Methods("GET")

	t.Run("get hq", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/BPKOPLPWXXX", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		hq, err := handlers.Decode[handlers.HeadquartersDTO](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "UL. PULAWSKA 15  WARSZAWA, MAZOWIECKIE, 02-515", hq.Address)
		assert.Equal(t, "PKO BANK POLSKI S.A.", hq.Name)
		assert.Equal(t, "PL", hq.CountryISO2)
		assert.Equal(t, "POLAND", hq.CountryName)
		assert.Equal(t, true, hq.IsHeadquarter)
		assert.Equal(t, "BPKOPLPWXXX", hq.SwiftCode)
		assert.Len(t, hq.Branches, 2)
		branchSwiftCodes := []string{hq.Branches[0].SwiftCode, hq.Branches[1].SwiftCode}
		assert.Contains(t, branchSwiftCodes, "BPKOPLPWCSD")
		assert.Contains(t, branchSwiftCodes, "BPKOPLPWGDG")
		assert.False(t, hq.Branches[0].IsHeadquarter)
		assert.False(t, hq.Branches[1].IsHeadquarter)
	})

	t.Run("get branch", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/BPKOPLPWCSD", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		branch, err := handlers.Decode[handlers.BranchDTO](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "WARSZAWA, MAZOWIECKIE", branch.Address)
		assert.Equal(t, "PKO BANK POLSKI S.A.", branch.Name)
		assert.Equal(t, "PL", branch.CountryISO2)
		assert.Equal(t, "POLAND", branch.CountryName)
		assert.Equal(t, false, branch.IsHeadquarter)
		assert.Equal(t, "BPKOPLPWCSD", branch.SwiftCode)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/NOTFOUND123", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid swift code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/INVALID", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "swift code length must be 11 characters", errMsg.Message)
	})
}

func TestGetAllBankUnitsForCountry(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{countryISO2code}", handlers.GetAllBankUnitsForCountry(bankUnitRepo, countryRepo)).Methods("GET")

	t.Run("get all for country", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/PL", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := handlers.Decode[handlers.SwiftCodeForCountryResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "PL", response.CountryISO2)
		assert.Equal(t, "POLAND", response.CountryName)
		assert.Len(t, response.SwiftCodes, 3)
		var swiftCodes []string
		for _, bu := range response.SwiftCodes {
			swiftCodes = append(swiftCodes, bu.SwiftCode)
		}
		assert.Contains(t, swiftCodes, "BPKOPLPWXXX")
		assert.Contains(t, swiftCodes, "BPKOPLPWCSD")
		assert.Contains(t, swiftCodes, "BPKOPLPWGDG")
	})

	t.Run("country not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/XX", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "country not found", errMsg.Message)
	})

	t.Run("invalid country code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/INVALID", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "country ISO2 code length must be 2 characters", errMsg.Message)
	})

	t.Run("country that does not have any bank units", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/DE", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		response, err := handlers.Decode[handlers.SwiftCodeForCountryResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "DE", response.CountryISO2)
		assert.Equal(t, "GERMANY", response.CountryName)
		assert.Len(t, response.SwiftCodes, 0)
	})
}

func TestDeleteBankUnit(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{swiftCode}", handlers.DeleteBankUnit(bankUnitRepo)).Methods("DELETE")

	t.Run("delete existing bank unit", withCleanup(func(t *testing.T) {
		assert.Equal(t, 4, len(Must(bankUnitRepo.GetAll(context.Background()))))

		req := httptest.NewRequest("DELETE", "/BEFNBGS1XXX", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		resp, err := handlers.Decode[handlers.SuccessResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "bank unit deleted", resp.Message)

		swiftcode := Must(model.NewSwiftCode("BEFNBGS1XXX"))
		_, err = bankUnitRepo.GetBySwiftCode(context.Background(), swiftcode)
		assert.ErrorIs(t, err, repo.ErrNotFound)
	}))

	t.Run("delete non-existing bank unit", withCleanup(func(t *testing.T) {
		assert.Equal(t, 4, len(Must(bankUnitRepo.GetAll(context.Background()))))

		req := httptest.NewRequest("DELETE", "/NOTEXISTXXX", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}))

	t.Run("invalid swift code", withCleanup(func(t *testing.T) {

		req := httptest.NewRequest("DELETE", "/INVALID", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "swift code length must be 11 characters", errMsg.Message)
	}))
}

func TestCreateBankUnit(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.CreateBankUnit(bankUnitRepo, countryRepo)).Methods("POST")

	t.Run("create valid bank unit", withCleanup(func(t *testing.T) {
		body := strings.NewReader(`{
			"swiftCode": "DEUTDEFFXXX",
			"countryISO2": "DE",
			"countryName": "GERMANY",
			"address": "TAUNUSANLAGE 12 FRANKFURT AM MAIN, HESSEN, 60325",
			"bankName": "DEUTSCHE BANK AG",
			"isHeadquarter": true
		}`)

		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		resp, err := handlers.Decode[handlers.SuccessResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "bank unit created", resp.Message)

		Must(model.NewSwiftCode("DEUTDEFFXXX"))
		swiftcode := Must(model.NewSwiftCode("DEUTDEFFXXX"))
		bankUnit, err := bankUnitRepo.GetBySwiftCode(context.Background(), swiftcode)
		assert.NoError(t, err)
		assert.Equal(t, "DEUTDEFFXXX", bankUnit.SwiftCode.String())
	}))

	t.Run("create duplicate bank unit", withCleanup(func(t *testing.T) {

		body := strings.NewReader(`{
			"swiftCode": "BPKOPLPWXXX",
			"countryISO2": "PL",
			"countryName": "POLAND",
			"address": "WARSZAWA",
			"bankName": "PKO BP",
			"isHeadquarter": true
		}`)

		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "duplicate swift code", errMsg.Message)
	}))

	t.Run("create with non-existing country", withCleanup(func(t *testing.T) {
		body := strings.NewReader(`{
			"swiftCode": "ABCDCNXXXXX",
			"countryISO2": "CN",
			"countryName": "CHINA",
			"address": "BEIJING",
			"bankName": "BANK OF CHINA",
			"isHeadquarter": true
		}`)

		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "country does not exist, make sure ISO2 code is matching with the name", errMsg.Message)
	}))

	t.Run("invalid request body", withCleanup(func(t *testing.T) {
		body := strings.NewReader(`invalid json`)

		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "invalid json data", errMsg.Message)
	}))
}

func Must[T any](value T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func exitIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func withCleanup(fn func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Cleanup(func() { resetTestData(context.Background()) })

		fn(t)
	}
}
