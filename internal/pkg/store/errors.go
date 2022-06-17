package store

import (
	"fmt"

	"github.com/google/uuid"
)

type NotFoundError struct {
	// Source defines an entity source (bucket name, table and etc.)
	Source string
	ID     uuid.UUID
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("entity with id %s from %q is not found", err.ID, err.Source)
}

type AlreadyExistError struct {
	// Source defines an entity source (bucket name, table and etc.)
	Source string
	ID     uuid.UUID
}

func (err AlreadyExistError) Error() string {
	return fmt.Sprintf("entity with id %s from %q already exist", err.ID, err.Source)
}
