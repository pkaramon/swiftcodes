package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/repo"
)

type branchDTO struct {
	Address       string `json:"address"`
	Name          string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

type headquartersDTO struct {
	branchDTO
	Branches []branchDTO `json:"branches"`
}

func branchToDTO(bu *model.BankUnit) branchDTO {
	return branchDTO{
		Address:       bu.Address,
		Name:          bu.Name,
		CountryISO2:   bu.Country.Code.String(),
		CountryName:   bu.Country.Name,
		IsHeadquarter: bu.IsHeadquarter,
		SwiftCode:     bu.SwiftCode.String(),
	}
}

func headquartersToDTO(hq *model.BankUnit, branches []*model.BankUnit) headquartersDTO {
	dto := headquartersDTO{
		branchDTO: branchToDTO(hq),
		Branches:  branchesToDTOS(branches),
	}
	return dto
}

func branchesToDTOS(bu []*model.BankUnit) []branchDTO {
	dtos := make([]branchDTO, len(bu))
	for i, b := range bu {
		dtos[i] = branchToDTO(b)
	}
	return dtos
}

func GetBankUnit(bankRepo repo.BankUnit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc := mux.Vars(r)["swiftCode"]
		swiftcode, err := model.NewSwiftCode(sc)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, err.Error())
			return
		}

		bankUnit, err := bankRepo.GetBySwiftCode(r.Context(), swiftcode)
		if err != nil {
			if err == repo.ErrNotFound {
				SendErrorMsg(w, http.StatusNotFound, "not found")
				return
			}
			SendErrorMsg(w, http.StatusInternalServerError, "server error")
		}

		if bankUnit.IsHeadquarter {
			branches, err := bankRepo.GetBranches(r.Context(), bankUnit.SwiftCode)
			if err != nil {
				SendServerError(w)
				return
			}
			dto := headquartersToDTO(bankUnit, branches)
			Encode(w, http.StatusOK, dto)
		} else {
			Encode(w, http.StatusOK, branchToDTO(bankUnit))
		}

	}
}

func GetAllBankUnitsForCountry(bankRepo repo.BankUnit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		countryISO2 := mux.Vars(r)["countryISO2code"]
		country, err := model.NewCountryISO2(countryISO2)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, err.Error())
			return
		}

		bankUnits, err := bankRepo.GetAllByCountry(r.Context(), country)
		if err != nil {
			SendServerError(w)
			return
		}

		Encode(w, http.StatusOK, branchesToDTOS(bankUnits))
	}
}

func DeleteBankUnit(bankRepo repo.BankUnit) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc := mux.Vars(r)["swiftCode"]
		swiftcode, err := model.NewSwiftCode(sc)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := bankRepo.Delete(r.Context(), swiftcode); err != nil {
			SendServerError(w)
			return
		}

		SendSuccessMsg(w, http.StatusOK, "bank unit deleted")
	}
}

func CreateBankUnit(bankRepo repo.BankUnit, countryRepo repo.Country) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := Decode[branchDTO](r)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, "invalid json data")
			return
		}
		country, err := model.NewCountry(data.CountryISO2, data.CountryName)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, err.Error())
			return
		}
		exists, err := countryRepo.Exists(r.Context(), country)
		if err != nil {
			log.Println(err)
			SendServerError(w)
			return
		}

		if !exists {
			SendErrorMsg(w, http.StatusBadRequest, "country does not exist, make sure ISO2 code is matching with the name")
			return
		}

		bu, err := model.NewBankUnit(
			data.SwiftCode,
			data.CountryISO2,
			data.CountryName,
			data.Address,
			data.Name,
			data.IsHeadquarter,
		)
		if err != nil {
			SendErrorMsg(w, http.StatusBadRequest, err.Error())
			return
		}

		err = bankRepo.Create(r.Context(), bu)
		if errors.Is(err, repo.ErrDuplicate) {
			SendErrorMsg(w, http.StatusConflict, "duplicate swift code")
			return
		}
		if err != nil {
			SendServerError(w)
			return
		}

		SendSuccessMsg(w, http.StatusCreated, "bank unit created")
	}
}
