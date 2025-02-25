package model

import (
	"errors"
	"fmt"
	"strings"
)

type BankUnit struct {
	SwiftCode     SwiftCode
	CountryISO2   CountryISO2
	Address       string
	Name          string
	IsHeadquarter bool
}

func NewBankUnit(swiftCode string, countryISO2 string, address, name string, isHeadquarter bool) (*BankUnit, error) {
	codeISO2, err := NewCountryISO2(countryISO2)
	if err != nil {
		return nil, err
	}

	sc, err := NewSwiftCode(swiftCode)
	if err != nil {
		return nil, err
	}

	if sc.CountryISO2() != codeISO2.code {
		return nil, errors.New("swift code and country ISO2 code mismatch")
	}

	if isHeadquarter && sc.BranchCode() != "XXX" {
		return nil, errors.New("headquarter must have branch code XXX")
	}

	if len(name) == 0 {
		return nil, errors.New("name cannot be empty")
	}

	return &BankUnit{
		SwiftCode:     sc,
		CountryISO2:   codeISO2,
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
