package model

import (
	"errors"
	"fmt"
	"strings"
)

type BankUnit struct {
	SwiftCode     SwiftCode
	Country       Country
	Address       string
	Name          string
	IsHeadquarter bool
}

func NewBankUnit(
	swiftCode string,
	countryISO2 string,
	countryName string,
	address,
	name string,
	isHeadquarter bool,
) (*BankUnit, error) {
	country, err := NewCountry(countryISO2, countryName)
	if err != nil {
		return nil, err
	}

	swiftcode, err := NewSwiftCode(swiftCode)
	if err != nil {
		return nil, err
	}

	if swiftcode.CountryISO2() != country.Code.String() {
		return nil, errors.New("swift code and country ISO2 code mismatch")
	}

	if isHeadquarter && swiftcode.BranchCode() != "XXX" {
		return nil, errors.New("headquarter must have branch code XXX")
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be empty")
	}

	return &BankUnit{
		SwiftCode:     swiftcode,
		Country:       country,
		Address:       address,
		Name:          name,
		IsHeadquarter: isHeadquarter,
	}, nil
}

type SwiftCode struct {
	s string
}

func NewSwiftCode(s string) (SwiftCode, error) {
	if len(s) != 11 {
		return SwiftCode{}, fmt.Errorf("swift code length must be 11 characters")
	}
	return SwiftCode{s: strings.ToUpper(s)}, nil
}

func (s SwiftCode) CountryISO2() string {
	return s.s[4:6]
}

func (s SwiftCode) BranchCode() string {
	return s.s[8:11]
}

func (s SwiftCode) BaseCode() string {
	return s.s[0:8]
}

func (s SwiftCode) HasHeadQuartersBranchCode() bool {
	return s.BranchCode() == "XXX"
}

func (s SwiftCode) String() string {
	return s.s
}
