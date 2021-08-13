// Package errors provides functions similar to functions in github.com/pkg/errors
// but uses fmt.Errorf function to wrap errors
package errors

import (
	"errors"
	"fmt"
)

func New(text string) error {
	return errors.New(text)
}

func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Wrap wraps an error using fmt.Errorf function. If err is nil, Wrap returns nil.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf wraps an error using fmt.Errorf function. If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
