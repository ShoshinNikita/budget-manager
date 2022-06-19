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
	require.True(t, IsNotFound(err))

	err = NewAlreadyExistError(Account{}, uuid.New())
	require.True(t, IsAlreadyExist(err))

	err = NewUserError(errors.New("qwerty"))
	require.True(t, IsUserError(err))
}
