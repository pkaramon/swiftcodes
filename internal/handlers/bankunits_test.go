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

	m.Run()
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
		assert.Equal(t, "swift code length must be 11 characters", errMsg.Error)
	})
}

func TestGetAllBankUnitsForCountry(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{countryISO2code}", handlers.GetAllBankUnitsForCountry(bankUnitRepo)).Methods("GET")

	t.Run("get all for country", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/PL", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		bankUnits, err := handlers.Decode[[]handlers.BranchDTO](rec.Result().Body)
		assert.Nil(t, err)
		assert.Len(t, bankUnits, 3)
		swiftCodes := []string{bankUnits[0].SwiftCode, bankUnits[1].SwiftCode, bankUnits[2].SwiftCode}
		assert.Contains(t, swiftCodes, "BPKOPLPWXXX")
		assert.Contains(t, swiftCodes, "BPKOPLPWCSD")
		assert.Contains(t, swiftCodes, "BPKOPLPWGDG")
	})

	t.Run("invalid country code", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/INVALID", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "country ISO2 code length must be 2 characters", errMsg.Error)
	})

	t.Run("country that does not have any bank units", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/DE", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		bankUnits, err := handlers.Decode[[]handlers.BranchDTO](rec.Result().Body)
		assert.Nil(t, err)
		assert.Len(t, bankUnits, 0)

	})
}

func TestDeleteBankUnit(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{swiftCode}", handlers.DeleteBankUnit(bankUnitRepo)).Methods("DELETE")

	t.Run("delete existing bank unit", func(t *testing.T) {
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
	})

	t.Run("delete non-existing bank unit", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/NOTEXISTXXX", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid swift code", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/INVALID", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "swift code length must be 11 characters", errMsg.Error)
	})
}

func TestCreateBankUnit(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.CreateBankUnit(bankUnitRepo, countryRepo)).Methods("POST")

	t.Run("create valid bank unit", func(t *testing.T) {
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
	})

	t.Run("create duplicate bank unit", func(t *testing.T) {
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
		assert.Equal(t, "duplicate swift code", errMsg.Error)
	})

	t.Run("create with non-existing country", func(t *testing.T) {
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
		assert.Equal(t, "country does not exist, make sure ISO2 code is matching with the name", errMsg.Error)
	})

	t.Run("invalid request body", func(t *testing.T) {
		body := strings.NewReader(`invalid json`)

		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		errMsg, err := handlers.Decode[handlers.ErrorResponse](rec.Result().Body)
		assert.Nil(t, err)
		assert.Equal(t, "invalid json data", errMsg.Error)
	})
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
