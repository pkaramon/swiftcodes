package entity

import (
	"errors"
	"fmt"
	"strings"
)

type BankUnit struct {
	SwiftCode     SwiftCode
	Country       *Country
	CountryName   string
	Address       string
	Name          string
	IsHeadquarter bool
}

func NewBankUnit(swiftCode string, countryISO2 string, countryName, address, name string, isHeadquarter bool) (*BankUnit, error) {
	sc, err := NewSwiftCode(swiftCode)
	if err != nil {
		return nil, err
	}
	country, err := NewCountry(countryISO2, countryName)
	if err != nil {
		return nil, err
	}

	if sc.CountryISO2() != country.CodeISO2 {
		return nil, errors.New("swift code and country ISO2 code mismatch")
	}

	if isHeadquarter && sc.BranchCode() != "XXX" {
		return nil, errors.New("headquarter must have branch code XXX")
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be empty")
	}

	if len(countryName) == 0 {
		return nil, errors.New("country name cannot be empty")
	}

	return &BankUnit{
		SwiftCode:     sc,
		Country:       country,
		CountryName:   strings.ToUpper(countryName),
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

type Country struct {
	CodeISO2 string
	Name     string
}

func NewCountry(codeISO2 string, name string) (*Country, error) {
	if len(codeISO2) != 2 {
		return nil, fmt.Errorf("country ISO2 code length must be 2 characters")
	}
	if len(name) == 0 {
		return nil, fmt.Errorf("country name cannot be empty")
	}

	return &Country{
		CodeISO2: strings.ToUpper(codeISO2),
		Name:     strings.ToUpper(name),
	}, nil
}
