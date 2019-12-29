package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorType int

const (
	UndefinedError ErrorType = iota
	UserError
	AppError
)

const DefaultErrorMessage = "unknown error"

type fundamental struct {
	originalError     error
	showOriginalError bool
	//
	msg string
	//
	errorType ErrorType
}

// Error implements 'error' interface.
func (err fundamental) Error() string {
	res := err.msg
	if err.showOriginalError {
		if res != "" {
			res += ": "
		}
		res += err.originalError.Error()
	}

	if res == "" {
		res = DefaultErrorMessage
	}
	return res
}

// New creates a new error with passed message
func New(msg string, opts ...Option) error {
	err := fundamental{
		originalError: errors.New(msg),
	}
	for _, opt := range opts {
		opt(&err)
	}

	return err
}

// Wrap wraps passed error and applies options
func Wrap(err error, opts ...Option) error {
	wrappedError := fundamental{originalError: err}
	if e, ok := err.(fundamental); ok {
		wrappedError = e
	}

	for _, opt := range opts {
		opt(&wrappedError)
	}

	return wrappedError
}

// GetOriginalError returns original error. If err isn't an instance of 'fundamental' type,
// passed error will be returned
func GetOriginalError(err error) error {
	if err, ok := err.(fundamental); ok {
		return err.originalError
	}
	return err
}

func GetErrorType(err error) (t ErrorType, ok bool) {
	if err, ok := err.(fundamental); ok {
		return err.errorType, true
	}
	return UndefinedError, false
}

// --------------------------------------------------
// Options
// --------------------------------------------------

type Option func(*fundamental)

// WithType sets passed Error Type
func WithType(t ErrorType) Option {
	return func(err *fundamental) {
		err.errorType = t
	}
}

// WithTypeIfNotSet sets passed Error Type if current Error Type is UndefinedError
func WithTypeIfNotSet(t ErrorType) Option {
	return func(err *fundamental) {
		if err.errorType == UndefinedError {
			err.errorType = t
		}
	}
}

// WithMsg sets message. It supports formatting
func WithMsg(msg string, a ...interface{}) Option {
	return func(err *fundamental) {
		if len(a) == 0 {
			err.msg = msg
			return
		}

		err.msg = fmt.Sprintf(msg, a...)
	}
}

// WithOriginalError exposes original error for printing
func WithOriginalError() Option {
	return func(err *fundamental) {
		err.showOriginalError = true
	}
}
