package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/web"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestBasicUsage(t *testing.T) {
	t.Parallel()

	cfg := app.Config{
		DBType:     "postgres",
		PostgresDB: pg.Config{Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
		Server:     web.Config{UseEmbed: true, SkipAuth: true, Credentials: nil, EnableProfiling: false},
	}
	prepareApp(t, &cfg, StartPostgreSQL)

	host := fmt.Sprintf("localhost:%d", cfg.Server.Port)

	for _, tt := range []struct {
		name string
		f    func(t *testing.T, host string)
	}{
		{name: "spend types", f: testBasicUsage_SpendTypes},
		{name: "incomes", f: testBasicUsage_Incomes},
		{name: "monthly payments", f: testBasicUsage_MonthlyPayments},
		{name: "spends", f: testBasicUsage_Spends},
		{name: "search spends", f: testBasicUsage_SearchSpends},
	} {
		tt := tt
		ok := t.Run(tt.name, func(t *testing.T) {
			tt.f(t, host)
		})
		if !ok {
			t.FailNow()
		}
	}
}

func testBasicUsage_SpendTypes(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []RequestCreated{
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "f00d"}},                  // 1
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "fastfood", ParentID: 1}}, // 2
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "pizza", ParentID: 2}},    // 3
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "travel"}},                // 4
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "avia", ParentID: 4}},     // 5
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "house"}},                 // 6
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "entertainment"}},         // 7
	} {
		var resp models.AddSpendTypeResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []RequestOK{
		{PUT, SpendTypesPath, models.EditSpendTypeReq{ID: 1, Name: ptrStr("food")}},
		{PUT, SpendTypesPath, models.EditSpendTypeReq{ID: 3, ParentID: ptrUint(1)}},
		{DELETE, SpendTypesPath, models.RemoveSpendTypeReq{ID: 5}},
	} {
		req.Send(t, host, nil)
	}

	// Check
	var resp models.GetSpendTypesResp
	RequestOK{GET, SpendTypesPath, nil}.Send(t, host, &resp)
	require.Equal(
		[]db.SpendType{
			{ID: 1, Name: "food"},
			{ID: 2, Name: "fastfood", ParentID: 1},
			{ID: 3, Name: "pizza", ParentID: 1},
			{ID: 4, Name: "travel"},
			{ID: 6, Name: "house"},
			{ID: 7, Name: "entertainment"},
		},
		resp.SpendTypes,
	)
}

func testBasicUsage_Incomes(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []RequestCreated{
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "salary", Income: 2500}},                // 1
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "gifts", Income: 500}},                  // 2
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "temp", Income: 100}},                   // 3
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "cashback", Notes: "123", Income: 100}}, // 4
	} {
		var resp models.AddIncomeResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []RequestOK{
		{PUT, IncomesPath, models.EditIncomeReq{ID: 2, Title: ptrStr("gift"), Notes: ptrStr("from friends")}},
		{PUT, IncomesPath, models.EditIncomeReq{ID: 4, Income: ptrFloat(50)}},
		{DELETE, IncomesPath, models.RemoveSpendTypeReq{ID: 3}},
	} {
		req.Send(t, host, nil)
	}

	// Check
	month := getCurrentMonth(t, host)

	expectedIncomes := []db.Income{
		{ID: 1, Title: "salary", Income: money.FromInt(2500)},
		{ID: 2, Title: "gift", Notes: "from friends", Income: money.FromInt(500)},
		{ID: 4, Title: "cashback", Notes: "123", Income: money.FromInt(50)},
	}
	for i := range expectedIncomes {
		expectedIncomes[i].Year = month.Year
		expectedIncomes[i].Month = month.Month
	}
	require.Equal(expectedIncomes, month.Incomes)

	checkMonth(require, 3050, 0, 0, month)
}

func testBasicUsage_MonthlyPayments(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []RequestCreated{
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "rent", TypeID: 6, Cost: 800}},         // 1
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "patre0n", Cost: 50}},                  // 2
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "netflix", Notes: "remove", Cost: 20}}, // 3
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "temp", Notes: "123", Cost: 100}},      // 4
	} {
		var resp models.AddMonthlyPaymentResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []RequestOK{
		{PUT, MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 2, Title: ptrStr("patreon"), Notes: ptrStr("with VAT")}},
		{PUT, MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 3, TypeID: ptrUint(7), Notes: ptrStr(""), Cost: ptrFloat(30)}},
		{DELETE, MonthlyPaymentsPath, models.RemoveMonthlyPaymentReq{ID: 4}},
	} {
		req.Send(t, host, nil)
	}

	// Check
	month := getCurrentMonth(t, host)

	expectedMonthlyPayments := []db.MonthlyPayment{
		{ID: 1, Title: "rent", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(800)},
		{ID: 2, Title: "patreon", Notes: "with VAT", Cost: money.FromInt(50)},
		{ID: 3, Title: "netflix", Type: &db.SpendType{ID: 7, Name: "entertainment"}, Cost: money.FromInt(30)},
	}
	for i := range expectedMonthlyPayments {
		expectedMonthlyPayments[i].Year = month.Year
		expectedMonthlyPayments[i].Month = month.Month
	}
	require.Equal(expectedMonthlyPayments, month.MonthlyPayments)

	checkMonth(require, 3050, -880, 0, month)
}

func testBasicUsage_Spends(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []RequestCreated{
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "bread", Notes: "fresh", TypeID: 1, Cost: 2}}, // 1
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "grocery", TypeID: 1, Cost: 10}},              // 2
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "milk", TypeID: 1, Cost: 2}},                  // 3
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 3, Title: "oil", TypeID: 1, Cost: 7}},            // 4
		{POST, SpendsPath, models.AddSpendReq{DayID: 3, Title: "dinner in KFC", TypeID: 2, Cost: 15}}, // 5
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 10, Title: "bicycle", Notes: "https://example.com", Cost: 500}}, // 6
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 11, Title: "meat", TypeID: 1, Cost: 20}}, // 7
		{POST, SpendsPath, models.AddSpendReq{DayID: 11, Title: "egg", TypeID: 1, Cost: 7}},   // 8
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 12, Title: "pizza", TypeID: 3, Cost: 100}}, // 9
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 15, Title: "book American Gods", Notes: "as a gift", Cost: 30}}, // 10
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "new mirror in the bathroom", TypeID: 6, Cost: 150}}, // 11
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "new towels", TypeID: 6, Cost: 50}},                  // 12
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "temp", Cost: 0}},                                    // 13
	} {
		var resp models.AddMonthlyPaymentResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []RequestOK{
		{PUT, SpendsPath, models.EditSpendReq{ID: 4, TypeID: ptrUint(0)}},
		{PUT, SpendsPath, models.EditSpendReq{ID: 8, Title: ptrStr("eggs"), Notes: ptrStr("10 count"), Cost: ptrFloat(8)}},
		{DELETE, SpendsPath, models.RemoveSpendReq{ID: 13}},
	} {
		req.Send(t, host, nil)
	}

	// Check
	month := getCurrentMonth(t, host)

	expectedDays := []db.Day{
		{ID: 1, Spends: []db.Spend{
			{ID: 1, Title: "bread", Type: &db.SpendType{ID: 1, Name: "food"}, Notes: "fresh", Cost: money.FromInt(2)},
			{ID: 2, Title: "grocery", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(10)},
			{ID: 3, Title: "milk", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(2)},
		}},
		{ID: 2, Spends: []db.Spend{}},
		{ID: 3, Spends: []db.Spend{
			{ID: 4, Title: "oil", Cost: money.FromInt(7)},
			{ID: 5, Title: "dinner in KFC", Type: &db.SpendType{ID: 2, Name: "fastfood", ParentID: 1}, Cost: money.FromInt(15)},
		}},
		{ID: 4, Spends: []db.Spend{}},
		{ID: 5, Spends: []db.Spend{}},
		{ID: 6, Spends: []db.Spend{}},
		{ID: 7, Spends: []db.Spend{}},
		{ID: 8, Spends: []db.Spend{}},
		{ID: 9, Spends: []db.Spend{}},
		{ID: 10, Spends: []db.Spend{
			{ID: 6, Title: "bicycle", Notes: "https://example.com", Cost: money.FromInt(500)},
		}},
		{ID: 11, Spends: []db.Spend{
			{ID: 7, Title: "meat", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(20)},
			{ID: 8, Title: "eggs", Type: &db.SpendType{ID: 1, Name: "food"}, Notes: "10 count", Cost: money.FromInt(8)},
		}},
		{ID: 12, Spends: []db.Spend{
			{ID: 9, Title: "pizza", Type: &db.SpendType{ID: 3, Name: "pizza", ParentID: 1}, Cost: money.FromInt(100)},
		}},
		{ID: 13, Spends: []db.Spend{}},
		{ID: 14, Spends: []db.Spend{}},
		{ID: 15, Spends: []db.Spend{
			{ID: 10, Title: "book American Gods", Notes: "as a gift", Cost: money.FromInt(30)},
		}},
		{ID: 16, Spends: []db.Spend{
			{ID: 11, Title: "new mirror in the bathroom", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(150)},
			{ID: 12, Title: "new towels", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(50)},
		}},
	}
	for i := len(expectedDays) + 1; i <= len(month.Days); i++ {
		expectedDays = append(expectedDays, db.Day{ID: uint(i), Spends: []db.Spend{}})
	}
	for i := range expectedDays {
		expectedDays[i].Year = month.Year
		expectedDays[i].Month = month.Month
		expectedDays[i].Day = int(expectedDays[i].ID)
		for j := range expectedDays[i].Spends {
			expectedDays[i].Spends[j].Year = expectedDays[i].Year
			expectedDays[i].Spends[j].Month = expectedDays[i].Month
			expectedDays[i].Spends[j].Day = expectedDays[i].Day
		}

		var prevSaldo money.Money
		if i > 0 {
			prevSaldo = expectedDays[i-1].Saldo
		}
		expectedDays[i].Saldo = prevSaldo.Add(month.DailyBudget)
		for _, s := range expectedDays[i].Spends {
			expectedDays[i].Saldo = expectedDays[i].Saldo.Sub(s.Cost)
		}
	}
	require.Equal(expectedDays, month.Days)

	checkMonth(require, 3050, -880, -894, month)
}

func testBasicUsage_SearchSpends(t *testing.T, host string) {
	// Prepare spends
	allSpends := []db.Spend{
		{ID: 1, Day: 1, Title: "bread", Type: &db.SpendType{ID: 1, Name: "food"}, Notes: "fresh", Cost: money.FromInt(2)},
		{ID: 2, Day: 1, Title: "grocery", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(10)},
		{ID: 3, Day: 1, Title: "milk", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(2)},
		{ID: 4, Day: 3, Title: "oil", Cost: money.FromInt(7)},
		{ID: 5, Day: 3, Title: "dinner in KFC", Type: &db.SpendType{ID: 2, Name: "fastfood", ParentID: 1}, Cost: money.FromInt(15)},
		{ID: 6, Day: 10, Title: "bicycle", Notes: "https://example.com", Cost: money.FromInt(500)},
		{ID: 7, Day: 11, Title: "meat", Type: &db.SpendType{ID: 1, Name: "food"}, Cost: money.FromInt(20)},
		{ID: 8, Day: 11, Title: "eggs", Type: &db.SpendType{ID: 1, Name: "food"}, Notes: "10 count", Cost: money.FromInt(8)},
		{ID: 9, Day: 12, Title: "pizza", Type: &db.SpendType{ID: 3, Name: "pizza", ParentID: 1}, Cost: money.FromInt(100)},
		{ID: 10, Day: 15, Title: "book American Gods", Notes: "as a gift", Cost: money.FromInt(30)},
		{ID: 11, Day: 16, Title: "new mirror in the bathroom", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(150)},
		{ID: 12, Day: 16, Title: "new towels", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(50)},
	}
	month := getCurrentMonth(t, host)
	for i := range allSpends {
		allSpends[i].Year = month.Year
		allSpends[i].Month = month.Month
	}

	getSpends := func(ids ...uint) []db.Spend {
		res := make([]db.Spend, 0, len(ids))
		for _, id := range ids {
			for _, spend := range allSpends {
				if spend.ID == id {
					res = append(res, spend)
					break
				}
			}
		}
		return res
	}
	newDate := func(day int) time.Time {
		return time.Date(month.Year, month.Month, day, 0, 0, 0, 0, time.UTC)
	}

	for _, tt := range []struct {
		name string
		req  models.SearchSpendsReq
		ids  []uint
	}{
		{
			name: "all spends",
			req:  models.SearchSpendsReq{},
			ids:  []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			name: "filter by title",
			req:  models.SearchSpendsReq{Title: "il"},
			ids:  []uint{3, 4},
		},
		{
			name: "filter by type",
			req:  models.SearchSpendsReq{TypeIDs: []uint{1, 2}},
			ids:  []uint{1, 2, 3, 5, 7, 8},
		},
		{
			name: "filter by notes",
			req:  models.SearchSpendsReq{Notes: "gift"},
			ids:  []uint{10},
		},
		{
			name: "filter by notes (exactly)",
			req:  models.SearchSpendsReq{Notes: "gift", NotesExactly: true},
			ids:  []uint{},
		},
		{
			name: "filter by cost (min)",
			req:  models.SearchSpendsReq{MinCost: 200},
			ids:  []uint{6},
		},
		{
			name: "filter by cost (max)",
			req:  models.SearchSpendsReq{MaxCost: 7},
			ids:  []uint{1, 3, 4},
		},
		{
			name: "filter by cost (min and max)",
			req:  models.SearchSpendsReq{MinCost: 10, MaxCost: 30},
			ids:  []uint{2, 5, 7, 10},
		},
		{
			name: "filter by time (after)",
			req:  models.SearchSpendsReq{After: newDate(15)},
			ids:  []uint{10, 11, 12},
		},
		{
			name: "filter by time (before)",
			req:  models.SearchSpendsReq{Before: newDate(2)},
			ids:  []uint{1, 2, 3},
		},
		{
			name: "filter by time (after and before)",
			req:  models.SearchSpendsReq{After: newDate(2), Before: newDate(7)},
			ids:  []uint{4, 5},
		},
		{
			name: "sort by cost desc",
			req:  models.SearchSpendsReq{MinCost: 7, MaxCost: 15, Sort: "cost", Order: "desc"},
			ids:  []uint{5, 2, 8, 4},
		},
		{
			name: "sort by title asc",
			req:  models.SearchSpendsReq{TypeIDs: []uint{1}, Sort: "title"},
			ids:  []uint{1, 8, 2, 7, 3},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			var resp models.SearchSpendsResp
			RequestOK{GET, SearchSpendsPath, tt.req}.Send(t, host, &resp)
			require.Equal(getSpends(tt.ids...), resp.Spends)
		})
	}
}

func getCurrentMonth(t *testing.T, host string) db.Month {
	// It's very unlikely that tests were run at the last seconds of a month,
	// and time.Now() will return the next month
	year, month, _ := time.Now().Date()

	var resp models.GetMonthResp
	RequestOK{GET, MonthsPath, models.GetMonthByDateReq{Year: year, Month: month}}.Send(t, host, &resp)
	return resp.Month
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
