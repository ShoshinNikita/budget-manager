package validator

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		input     string
		tagName   string
		s         any
		wantError string
	}{
		{
			input: `{"status":"old"}`,
			s: &struct {
				Status Status `json:"status"`
			}{},
			wantError: "",
		},
		{
			input: `{"status":"old"}`,
			s: &struct {
				Status Valid[Status] `json:"status"`
			}{},
			wantError: `invalid field "status": only "new" status is allowed`,
		},
		{
			input: `{"User": {"id": ""}}`,
			s: &struct {
				User Valid[User]
			}{},
			wantError: `invalid field "User": user id can't be empty`,
		},
		{
			input: `{"user": {"id": "test"}}`,
			s: &struct {
				User Valid[User] `json:"user"`
			}{},
			wantError: "",
		},
		{
			input:   `{"Nested": {"Nested": {"status": "old"}}}`,
			tagName: "form",
			s: &struct {
				Nested struct {
					Nested *struct {
						Status Valid[Status] `json:"status" form:"_status"`
					}
				}
			}{},
			wantError: `invalid field "_status": only "new" status is allowed`,
		},
	} {
		err := json.Unmarshal([]byte(tt.input), tt.s)
		require.NoError(t, err)

		v := NewValidator()
		if tt.tagName != "" {
			v.SetTagName(tt.tagName)
		}

		err = v.Validate(tt.s)
		if tt.wantError == "" {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, tt.wantError)
		}
	}
}

type Status string

func (s Status) IsValid() error {
	if s != "new" {
		return errors.New(`only "new" status is allowed`)
	}
	return nil
}

type User struct {
	ID string
}

func (u User) IsValid() error {
	if u.ID == "" {
		return errors.New("user id can't be empty")
	}
	return nil
}
