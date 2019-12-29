package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		errorMsg  string
		options   []Option
		want      fundamental
		wantPrint string
	}{
		{
			errorMsg: "test",
			options: []Option{
				WithMsg("%s: %d", "test", 15),
				WithType(UserError),
			},
			want: fundamental{
				originalError:     errors.New("test"),
				showOriginalError: false,
				msg:               "test: 15",
				errorType:         UserError,
			},
			wantPrint: "test: 15",
		},
		{
			errorMsg: "test",
			options: []Option{
				WithTypeIfNotSet(UserError),
			},
			want: fundamental{
				originalError:     errors.New("test"),
				showOriginalError: false,
				msg:               "",
				errorType:         UserError,
			},
			wantPrint: DefaultErrorMessage,
		},
		{
			errorMsg: "test",
			options: []Option{
				WithOriginalError(),
				WithMsg("123"),
				WithMsg("456"),
				WithType(UserError),
				WithTypeIfNotSet(AppError),
			},
			want: fundamental{
				originalError:     errors.New("test"),
				showOriginalError: true,
				msg:               "456",
				errorType:         UserError,
			},
			wantPrint: "456: test",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()

			require := require.New(t)

			// Check internal structure

			err := New(tt.errorMsg, tt.options...)
			// We have to check errors manually because addresses are different
			require.Equal(err.(fundamental).originalError.Error(), tt.want.originalError.Error())
			tt.want.originalError = err.(fundamental).originalError
			require.Equal(tt.want, err)

			// Check 'Error' method

			require.Equal(tt.wantPrint, fmt.Sprint(err))

			// Check 'Get...' functions

			errType, ok := GetErrorType(err)
			require.True(ok)
			require.Equal(tt.want.errorType, errType)

			require.Equal(tt.want.originalError, GetOriginalError(err))
		})
	}
}

func TestWrap(t *testing.T) {
	require := require.New(t)

	originalErr := errors.New("original error")

	err := Wrap(originalErr, WithOriginalError(), WithMsg("first wrap"))
	want := fundamental{
		originalError:     originalErr,
		showOriginalError: true,
		msg:               "first wrap",
		errorType:         UndefinedError,
	}
	require.Equal(want, err.(fundamental))

	err = Wrap(err, WithMsg("second wrap"), WithTypeIfNotSet(AppError))
	want.errorType = AppError
	want.msg = "second wrap"
	require.Equal(want, err.(fundamental))

	err = Wrap(err, WithMsg("third wrap"), WithTypeIfNotSet(UserError))
	want.msg = "third wrap"
	require.Equal(want, err.(fundamental))
}
