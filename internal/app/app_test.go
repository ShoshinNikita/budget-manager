package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCalculateTimeToNextMonthInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		now  time.Time
		want time.Duration
	}{
		{
			now:  time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			want: 31 * 24 * time.Hour,
		},
		{
			now:  time.Date(2021, time.January, 31, 22, 50, 5, 0, time.UTC),
			want: time.Hour + 9*time.Minute + 55*time.Second,
		},
		{
			now:  time.Date(2021, time.January, 31, 22, 50, 5, 0, time.UTC),
			want: time.Hour + 9*time.Minute + 55*time.Second,
		},
		{
			now:  time.Date(2021, time.April, 1, 0, 0, 45, 0, time.UTC),
			want: 29*24*time.Hour + 23*time.Hour + 59*time.Minute + 15*time.Second,
		},
		{
			now:  time.Date(2021, time.April, 30, 23, 59, 59, 0, time.UTC),
			want: time.Second,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := calculateTimeToNextMonthInit(tt.now)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.now.Month()+1, tt.now.Add(got).Month())
		})
	}
}
