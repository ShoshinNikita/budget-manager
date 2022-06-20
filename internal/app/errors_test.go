package app

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIsError(t *testing.T) {
	var err error

	err = NewNotFoundError(Account{}, uuid.New())
	require.NotNil(t, AsNotFound(err))

	err = NewAlreadyExistError(Account{}, uuid.New())
	require.NotNil(t, AsAlreadyExist(err))

	err = NewUserError(errors.New("qwerty"))
	require.NotNil(t, AsUserError(err))
}
