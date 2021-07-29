package tests

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func ptrStr(v string) *string     { return &v }
func ptrUint(v uint) *uint        { return &v }
func ptrFloat(v float64) *float64 { return &v }

func runSubtest(t *testing.T, name string, f func(*testing.T)) {
	ok := t.Run(name, func(t *testing.T) {
		f(t)
	})
	if !ok {
		t.FailNow()
	}
}

func checkMonth(require *require.Assertions, incomes, monthlyPayments, spends float64, month db.Month) {
	inc := money.FromFloat(incomes)
	require.Equal(inc, month.TotalIncome)

	mp := money.FromFloat(monthlyPayments)
	sp := money.FromFloat(spends)
	total := mp.Add(sp)
	require.Equal(total, month.TotalSpend)

	res := inc.Add(total) // spends and monthlyPayments are < 0
	require.Equal(res, month.Result)

	dailyBudget := inc.Add(mp).Div(int64(len(month.Days)))
	require.Equal(dailyBudget, month.DailyBudget)
}

func newRequest(t *testing.T, method Method, url string, body io.Reader) (req *http.Request, cancelCtx func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	req, err := http.NewRequestWithContext(ctx, string(method), url, body)
	require.NoError(t, err)

	return req, cancel
}
