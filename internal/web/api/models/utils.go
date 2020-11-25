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

// notPositiveFieldError must be used when field is negative or zero (<= 0)
func notPositiveFieldError(fieldName string) error {
	return errors.Errorf("%s must be greater than zero", fieldName)
}

// negativeFieldError must be used when field is negative (< 0)
func negativeFieldError(fieldName string) error {
	return errors.Errorf("%s must be greater or equal to zero", fieldName)
}
