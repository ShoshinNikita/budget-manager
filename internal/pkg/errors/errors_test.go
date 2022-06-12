package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	var (
		ErrNil error
		ErrNew = New("new error")
	)

	require.NotNil(Wrap(ErrNil, "wrap"))
	require.NotNil(Wrapf(ErrNil, "wrap %s", "1"))

	for _, tt := range []struct {
		err      error
		wantText string
	}{
		{ErrNew, "new error"},
		{Wrap(ErrNew, "wrap"), "wrap: new error"},
		{Wrap(Wrap(ErrNew, "wrap"), "wrap1"), "wrap1: wrap: new error"},
		{Wrapf(ErrNew, "format %d", 15), "format 15: new error"},
		{Errorf("some message: %w", ErrNew), "some message: new error"},
	} {
		got := Is(tt.err, ErrNew)
		require.True(got)
		require.Equal(tt.wantText, tt.err.Error())
	}
}
