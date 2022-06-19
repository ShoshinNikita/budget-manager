package app

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Entity interface {
	GetEntityName() string
}

type NotFoundError struct {
	EntityName string
	ID         uuid.UUID
}

func NewNotFoundError(e Entity, id uuid.UUID) *NotFoundError {
	return &NotFoundError{
		EntityName: e.GetEntityName(),
		ID:         id,
	}
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s is not found", err.EntityName, err.ID)
}

func IsNotFound(err error) bool {
	var notFoundError *NotFoundError
	return errors.As(err, &notFoundError)
}

type AlreadyExistError struct {
	EntityName string
	ID         uuid.UUID
}

func NewAlreadyExistError(e Entity, id uuid.UUID) *AlreadyExistError {
	return &AlreadyExistError{
		EntityName: e.GetEntityName(),
		ID:         id,
	}
}

func (err *AlreadyExistError) Error() string {
	return fmt.Sprintf("%s with id %s already exist", err.EntityName, err.ID)
}

func IsAlreadyExist(err error) bool {
	var alreadyExistError *AlreadyExistError
	return errors.As(err, &alreadyExistError)
}

type UserError struct {
	Err error
}

func NewUserError(err error) *UserError {
	return &UserError{
		Err: err,
	}
}

func (err *UserError) Error() string {
	return fmt.Sprintf("user error: %s", err.Err)
}

func (err *UserError) Unwrap() error {
	return err.Err
}

func IsUserError(err error) bool {
	var userError *UserError
	return errors.As(err, &userError)
}
