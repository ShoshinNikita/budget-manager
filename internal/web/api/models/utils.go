package models

import "github.com/pkg/errors"

// emptyFieldError must be used when field of type string is empty
func emptyFieldError(fieldName string) error {
	return errors.Errorf("%s can't be empty", fieldName)
}

// emptyOrZeroFieldError must be used when field of type int or float is empty or zero
func emptyOrZeroFieldError(fieldName string) error {
	return errors.Errorf("%s can't be empty or zero", fieldName)
}

// notPositiveFieldError must be used when field is negative or zero (< 1)
func notPositiveFieldError(fieldName string) error {
	return errors.Errorf("%s must be greater than zero", fieldName)
}
