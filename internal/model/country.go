package model

import (
	"errors"
	"fmt"
	"strings"
)

type Country struct {
	Code CountryISO2
	Name string
}

func NewCountry(code string, name string) (Country, error) {
	iso2, err := NewCountryISO2(code)

	if err != nil {
		return Country{}, err
	}

	if len(name) == 0 {
		return Country{}, errors.New("country name cannot be empty")
	}

	return Country{Code: iso2, Name: strings.ToUpper(name)}, nil
}

type CountryISO2 struct {
	code string
}

func NewCountryISO2(codeISO2 string) (CountryISO2, error) {
	if len(codeISO2) != 2 {
		return CountryISO2{}, fmt.Errorf("country ISO2 code length must be 2 characters")
	}

	return CountryISO2{
		code: strings.ToUpper(codeISO2),
	}, nil
}

func (c CountryISO2) String() string {
	return c.code
}
