package store

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type NotFoundError struct {
	EntityName string
	ID         uuid.UUID
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

func (err *AlreadyExistError) Error() string {
	return fmt.Sprintf("%s with id %s already exist", err.EntityName, err.ID)
}

func IsAlreadyExist(err error) bool {
	var alreadyExistError *AlreadyExistError
	return errors.As(err, &alreadyExistError)
}
