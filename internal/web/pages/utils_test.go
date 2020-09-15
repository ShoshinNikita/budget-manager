package pages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToShortMonth(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	tests := []struct {
		m    time.Month
		want string
	}{
		{m: time.January, want: "Jan"},
		{m: time.February, want: "Feb"},
		{m: time.March, want: "Mar"},
		{m: time.April, want: "Apr"},
		{m: time.May, want: "May"},
		{m: time.June, want: "June"},
		{m: time.July, want: "July"},
		{m: time.August, want: "Aug"},
		{m: time.September, want: "Sep"},
		{m: time.October, want: "Oct"},
		{m: time.November, want: "Nov"},
		{m: time.December, want: "Dec"},
	}

	for _, tt := range tests {
		res := toShortMonth(tt.m)
		require.Equal(tt.want, res)
	}
}
