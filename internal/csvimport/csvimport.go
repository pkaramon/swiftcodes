package csvimport

import (
	"context"
	"io"
	"log"

	"github.com/pkarmon/swiftcodes/internal/csvmapper"
	"github.com/pkarmon/swiftcodes/internal/model"
	"github.com/pkarmon/swiftcodes/internal/repo"
)

func Countries(ctx context.Context, src io.Reader, r repo.Country) error {
	mapper := csvmapper.New(src, []string{"Name", "Code"}, mapCSVRecordToCountry)

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

func mapCSVRecordToCountry(record []string) (model.Country, error) {
	country, err := model.NewCountry(record[1], record[0])
	if err != nil {
		return model.Country{}, err
	}
	return country, nil
}

func BankUnits(ctx context.Context, src io.Reader, r repo.BankUnit) error {
	mapper := csvmapper.New(src, []string{"SWIFT CODE", "COUNTRY ISO2 CODE", "COUNTRY NAME", "NAME", "ADDRESS"}, mapCSVRecordToBankUnit)

	bankUnits, err := mapper.MapAll()
	if err != nil {
		return err
	}

	log.Printf("Successfully loaded %d bank units", len(bankUnits))

	if err := r.BulkCreate(ctx, bankUnits); err != nil {
		return err
	}

	log.Println("Bank units imported successfully")

	return nil
}

func mapCSVRecordToBankUnit(record []string) (*model.BankUnit, error) {
	swiftCode, err := model.NewSwiftCode(record[0])
	if err != nil {
		return nil, err
	}

	countryISO2, err := model.NewCountryISO2(record[1])
	if err != nil {
		return nil, err
	}

	isHeadquarter := swiftCode.HasHeadQuartersBranchCode()

	bankUnit, err := model.NewBankUnit(
		swiftCode.String(),
		countryISO2.String(),
		record[2],
		record[4],
		record[3],
		isHeadquarter,
	)
	if err != nil {
		return nil, err
	}

	return bankUnit, nil
}
