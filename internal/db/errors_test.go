package db

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestErrorTypes(t *testing.T) {
	t.Run("plain error", func(t *testing.T) {
		require := require.New(t)
		var err error

		err = errors.New("some error")
		require.False(IsBadRequestError(err))
		require.False(IsInternalError(err))

		err = errors.Wrap(err, "wrap msg")
		require.False(IsBadRequestError(err))
		require.False(IsInternalError(err))
	})

	t.Run("internal error", func(t *testing.T) {
		require := require.New(t)
		var err error

		err = internalError(errors.New("internal error"))
		require.False(IsBadRequestError(err))
		require.True(IsInternalError(err))

		// Wrap with internalError
		err = internalError(err)
		require.False(IsBadRequestError(err))
		require.True(IsInternalError(err))

		// Wrap and unwrap with github.com/pkg/errors
		err = errors.Wrap(err, "wrap msg")
		require.False(IsBadRequestError(err))
		require.False(IsInternalError(err))

		err = errors.Cause(err)
		require.False(IsBadRequestError(err))
		require.True(IsInternalError(err))
	})

	t.Run("bad request error", func(t *testing.T) {
		require := require.New(t)
		var err error

		err = badRequestError(errors.New("bad request error"))
		require.True(IsBadRequestError(err))
		require.False(IsInternalError(err))

		// Wrap with badRequestError
		err = badRequestError(err)
		require.True(IsBadRequestError(err))
		require.False(IsInternalError(err))

		// Wrap and unwrap with github.com/pkg/errors
		err = errors.Wrap(err, "wrap msg")
		require.False(IsBadRequestError(err))
		require.False(IsInternalError(err))

		err = errors.Cause(err)
		require.True(IsBadRequestError(err))
		require.False(IsInternalError(err))
	})
}
