package money

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeCurrency(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal json", func(t *testing.T) {
		for _, tt := range []struct {
			input        string
			wantErr      bool
			wantCurrency Currency
		}{
			{input: `{"c":"BTC"}`, wantCurrency: "BTC"},
			{input: `{"c":"rUb"}`, wantCurrency: "RUB"},
			{input: `{"c":"test"}`, wantErr: true},
		} {
			var res struct {
				Currency Currency `json:"c"`
			}
			if err := json.Unmarshal([]byte(tt.input), &res); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantCurrency, res.Currency)
		}
	})
}
