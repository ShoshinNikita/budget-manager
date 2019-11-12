package db

import "github.com/pkg/errors"

// --------------------------------------------------
// Bad Request Error
// --------------------------------------------------

// IsBadRequestError checks whether error was caused by invalid request
func IsBadRequestError(err error) bool {
	_, ok := err.(badRequestErrorType)
	return ok
}

// badRequestErrorType is used for errors caused by invalid data (empty title, for example)
type badRequestErrorType struct {
	err error
}

func (err badRequestErrorType) Error() string {
	return err.err.Error()
}

// badRequestError wraps error in badRequestErrorType type
func badRequestError(err error) error {
	if err == nil {
		return nil
	}

	if brErr, ok := err.(badRequestErrorType); ok {
		return brErr
	}

	return badRequestErrorType{err: err}
}

// badRequestErrorWrap calls errors.Wrap and badRequestError
func badRequestErrorWrap(err error, msg string) error {
	err = errors.Wrap(err, msg)
	return badRequestError(err)
}

// badRequestErrorWrap calls errors.Wrapf and badRequestError
func badRequestErrorWrapf(err error, format string, args ...interface{}) error {
	err = errors.Wrapf(err, format, args...)
	return badRequestError(err)
}

// --------------------------------------------------
// Internal Error
// --------------------------------------------------

// IsBadRequestError checks whether error was caused by internal error
func IsInternalError(err error) bool {
	_, ok := err.(internalErrorType)
	return ok
}

// internalErrorType is used for internal errors (db error, for example)
type internalErrorType struct {
	err error
}

func (err internalErrorType) Error() string {
	return err.err.Error()
}

// badRequestError wraps error in internalErrorType type
func internalError(err error) error {
	if err == nil {
		return nil
	}

	if irErr, ok := err.(internalErrorType); ok {
		return irErr
	}

	return internalErrorType{err: err}
}

// internalErrorWrap calls errors.Wrap and internalError
func internalErrorWrap(err error, msg string) error {
	err = errors.Wrap(err, msg)
	return internalError(err)
}

// internalErrorWrap calls errors.Wrapf and internalError
func internalErrorWrapf(err error, format string, args ...interface{}) error {
	err = errors.Wrapf(err, format, args...)
	return internalError(err)
}

// --------------------------------------------------

// errorWrap respects type of the original error
func errorWrap(err error, msg string) error {
	switch err := err.(type) {
	case internalErrorType:
		return internalErrorWrap(err, msg)
	case badRequestErrorType:
		return badRequestErrorWrap(err, msg)
	default:
		return errors.Wrap(err, msg)
	}
}

// errorWrapf respects type of the original error
func errorWrapf(err error, format string, args ...interface{}) error {
	switch err := err.(type) {
	case internalErrorType:
		return internalErrorWrapf(err, format, args...)
	case badRequestErrorType:
		return badRequestErrorWrapf(err, format, args...)
	default:
		return errors.Wrapf(err, format, args...)
	}
}
