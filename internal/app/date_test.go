package app

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateDecode(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		textInput string
		wantError string
		wantDate  Date
	}{
		{
			textInput: "2022-06-22",
			wantDate:  Date{2022, time.June, 22},
		},
		{
			textInput: "2000-13-55",
			wantDate:  Date{2000, 13, 55},
		},
		{
			textInput: "",
			wantError: `invalid date "": expected 10 characters`,
		},
		{
			textInput: "20222-06-22",
			wantError: `invalid date "20222-06-22": expected 10 characters`,
		},
		{
			textInput: "cccc-06-22",
			wantError: `invalid date "cccc-06-22": invalid year`,
		},
		{
			textInput: "2022-A6-22",
			wantError: `invalid date "2022-A6-22": invalid month`,
		},
		{
			textInput: "2022-06-2&",
			wantError: `invalid date "2022-06-2&": invalid day`,
		},
	} {
		tt := tt
		t.Run("", func(t *testing.T) {
			require := require.New(t)

			checkErr := func(err error) {
				if tt.wantError == "" {
					require.NoError(err)
				} else {
					require.EqualError(err, tt.wantError)
				}
			}

			var textRes Date
			err := textRes.UnmarshalText([]byte(tt.textInput))
			checkErr(err)
			require.Equal(tt.wantDate, textRes)

			var jsonRes struct {
				Date Date `json:"date"`
			}
			err = json.Unmarshal([]byte(fmt.Sprintf(`{"date": "%s"}`, tt.textInput)), &jsonRes)
			checkErr(err)

			require.Equal(tt.wantDate, jsonRes.Date)
		})
	}
}

func TestDateEncode(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	type x struct {
		Date Date `json:"date"`
	}
	data, err := json.Marshal(x{Date{2022, time.January, 15}})
	require.NoError(err)
	require.Equal(`{"date":"2022-01-15"}`, string(data))

	require.Equal(`2000-01-29`, fmt.Sprint(Date{2000, time.January, 29}))
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		d       Date
		isValid bool
	}{
		{Date{2022, time.January, 25}, true},
		{Date{2022, 0, 25}, false},
		{Date{2022, time.April, 40}, false},
		{Date{2022, time.April, -5}, false},
		{Date{2022, 15, -5}, false},
	} {
		tt := tt
		t.Run("", func(t *testing.T) {
			err := tt.d.IsValid()
			if tt.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
