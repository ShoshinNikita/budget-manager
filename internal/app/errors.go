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

func AsNotFound(err error) (res *NotFoundError) {
	if errors.As(err, &res) {
		return res
	}
	return nil
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

func AsAlreadyExist(err error) (res *AlreadyExistError) {
	if errors.As(err, &res) {
		return res
	}
	return nil
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

func AsUserError(err error) (res *UserError) {
	if errors.As(err, &res) {
		return res
	}
	return nil
}
