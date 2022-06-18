package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsError(t *testing.T) {
	var err error = &NotFoundError{}
	require.True(t, IsNotFound(err))

	err = &AlreadyExistError{}
	require.True(t, IsAlreadyExist(err))
}
