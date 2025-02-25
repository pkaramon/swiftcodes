package entity

import (
	"errors"
	"strings"
)

type Country struct {
	Code CountryISO2
	Name string
}

func NewCountry(code CountryISO2, name string) (*Country, error) {
	if len(name) == 0 {
		return nil, errors.New("country name cannot be empty")
	}

	return &Country{Code: code, Name: strings.ToUpper(name)}, nil
}
