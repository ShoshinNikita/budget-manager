package money_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ShoshinNikita/budget-manager/v2/internal/pkg/money"
)

func TestEncodeMoney(t *testing.T) {
	t.Parallel()

	t.Run("json marshal", func(t *testing.T) {
		for _, tt := range []struct {
			input Money
			want  string
		}{
			{input: FromInt(357), want: `{"money":"357"}`},
			{input: FromFloat(154.3), want: `{"money":"154.3"}`},
			{input: FromFloat(0.07), want: `{"money":"0.07"}`},
			{input: FromFloat(15.07300001), want: `{"money":"15.07300001"}`},
			{input: FromFloat(15.078000008), want: `{"money":"15.078000008"}`},
		} {
			testStruct := struct {
				Money Money `json:"money"`
			}{tt.input}

			data, err := json.Marshal(testStruct)

			require.Nil(t, err)
			require.Equal(t, tt.want, string(data))
		}
	})

	t.Run("json unmarshal", func(t *testing.T) {
		for _, tt := range []struct {
			input   string
			want    float64
			wantErr bool
		}{
			{input: `{"money":357}`, want: 357},
			{input: `{"money":"154.30"}`, want: 154.30},
			{input: `{"money":0.070000001}`, want: 0.070000001},
			{input: `{"money":"test"}`, wantErr: true},
		} {
			var res struct {
				Money Money `json:"money"`
			}
			if err := json.Unmarshal([]byte(tt.input), &res); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, res.Money.Float())
		}
	})
}
