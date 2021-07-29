package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/web"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestBasicUsage(t *testing.T) {
	cfg := app.Config{
		DBType:     "postgres",
		PostgresDB: pg.Config{Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
		Server:     web.Config{UseEmbed: true, SkipAuth: true, Credentials: nil, EnableProfiling: false},
	}
	prepareApp(t, &cfg, startPostgreSQL)

	host := fmt.Sprintf("localhost:%d", cfg.Server.Port)

	runSubtest(t, "spend types", func(t *testing.T) {
		testBasicUsage_SpendTypes(t, host)
	})
	runSubtest(t, "incomes", func(t *testing.T) {
		testBasicUsage_Incomes(t, host)
	})
	runSubtest(t, "monthly payments", func(t *testing.T) {
		testBasicUsage_MonthlyPayments(t, host)
	})
	runSubtest(t, "spends", func(t *testing.T) {
		testBasicUsage_Spends(t, host)
	})
}

func testBasicUsage_SpendTypes(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []Request{
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "f00d"}, http.StatusCreated, ""},                  // 1
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "fastfood", ParentID: 1}, http.StatusCreated, ""}, // 2
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "pizza", ParentID: 2}, http.StatusCreated, ""},    // 3
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "travel"}, http.StatusCreated, ""},                // 4
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "avia", ParentID: 4}, http.StatusCreated, ""},     // 5
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "house"}, http.StatusCreated, ""},                 // 6
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "entertainment"}, http.StatusCreated, ""},         // 7
	} {
		var resp models.AddSpendTypeResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []Request{
		{PUT, SpendTypesPath, models.EditSpendTypeReq{ID: 1, Name: ptrStr("food")}, http.StatusOK, ""},
		{PUT, SpendTypesPath, models.EditSpendTypeReq{ID: 3, ParentID: ptrUint(1)}, http.StatusOK, ""},
		{DELETE, SpendTypesPath, models.RemoveSpendTypeReq{ID: 5}, http.StatusOK, ""},
	} {
		req.Send(t, host, nil)
	}

	// Check
	var resp models.GetSpendTypesResp
	Request{GET, SpendTypesPath, nil, http.StatusOK, ""}.Send(t, host, &resp)
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
	for i, req := range []Request{
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "salary", Income: 2500}, http.StatusCreated, ""},                // 1
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "gifts", Income: 500}, http.StatusCreated, ""},                  // 2
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "temp", Income: 100}, http.StatusCreated, ""},                   // 3
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "cashback", Notes: "123", Income: 100}, http.StatusCreated, ""}, // 4
	} {
		var resp models.AddIncomeResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []Request{
		{PUT, IncomesPath, models.EditIncomeReq{ID: 2, Title: ptrStr("gift"), Notes: ptrStr("from friends")}, http.StatusOK, ""},
		{PUT, IncomesPath, models.EditIncomeReq{ID: 4, Income: ptrFloat(50)}, http.StatusOK, ""},
		{DELETE, IncomesPath, models.RemoveSpendTypeReq{ID: 3}, http.StatusOK, ""},
	} {
		req.Send(t, host, nil)
	}

	// Check
	var resp models.GetMonthResp
	Request{GET, MonthsPath, models.GetMonthByIDReq{ID: 1}, http.StatusOK, ""}.Send(t, host, &resp)

	expectedIncomes := []db.Income{
		{ID: 1, Title: "salary", Income: money.FromInt(2500)},
		{ID: 2, Title: "gift", Notes: "from friends", Income: money.FromInt(500)},
		{ID: 4, Title: "cashback", Notes: "123", Income: money.FromInt(50)},
	}
	for i := range expectedIncomes {
		expectedIncomes[i].Year = resp.Month.Year
		expectedIncomes[i].Month = resp.Month.Month
	}
	require.Equal(expectedIncomes, resp.Month.Incomes)

	checkMonth(require, 3050, 0, 0, resp.Month)
}

func testBasicUsage_MonthlyPayments(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []Request{
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "rent", TypeID: 6, Cost: 800}, http.StatusCreated, ""},         // 1
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "patre0n", Cost: 50}, http.StatusCreated, ""},                  // 2
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "netflix", Notes: "remove", Cost: 20}, http.StatusCreated, ""}, // 3
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "temp", Notes: "123", Cost: 100}, http.StatusCreated, ""},      // 4
	} {
		var resp models.AddMonthlyPaymentResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []Request{
		{PUT, MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 2, Title: ptrStr("patreon"), Notes: ptrStr("with VAT")}, http.StatusOK, ""},
		{PUT, MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 3, TypeID: ptrUint(7), Notes: ptrStr(""), Cost: ptrFloat(30)}, http.StatusOK, ""},
		{DELETE, MonthlyPaymentsPath, models.RemoveMonthlyPaymentReq{ID: 4}, http.StatusOK, ""},
	} {
		req.Send(t, host, nil)
	}

	// Check
	var resp models.GetMonthResp
	Request{GET, MonthsPath, models.GetMonthByIDReq{ID: 1}, http.StatusOK, ""}.Send(t, host, &resp)

	expectedMonthlyPayments := []db.MonthlyPayment{
		{ID: 1, Title: "rent", Type: &db.SpendType{ID: 6, Name: "house"}, Cost: money.FromInt(800)},
		{ID: 2, Title: "patreon", Notes: "with VAT", Cost: money.FromInt(50)},
		{ID: 3, Title: "netflix", Type: &db.SpendType{ID: 7, Name: "entertainment"}, Cost: money.FromInt(30)},
	}
	for i := range expectedMonthlyPayments {
		expectedMonthlyPayments[i].Year = resp.Month.Year
		expectedMonthlyPayments[i].Month = resp.Month.Month
	}
	require.Equal(expectedMonthlyPayments, resp.Month.MonthlyPayments)

	checkMonth(require, 3050, -880, 0, resp.Month)
}

func testBasicUsage_Spends(t *testing.T, host string) {
	require := require.New(t)

	// Add
	for i, req := range []Request{
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "bread", Notes: "fresh", TypeID: 1, Cost: 2}, http.StatusCreated, ""}, // 1
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "grocery", TypeID: 1, Cost: 10}, http.StatusCreated, ""},              // 2
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "milk", TypeID: 1, Cost: 2}, http.StatusCreated, ""},                  // 3
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 3, Title: "oil", TypeID: 1, Cost: 7}, http.StatusCreated, ""},            // 4
		{POST, SpendsPath, models.AddSpendReq{DayID: 3, Title: "dinner in KFC", TypeID: 2, Cost: 15}, http.StatusCreated, ""}, // 5
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 10, Title: "bicycle", Notes: "https://example.com", Cost: 500}, http.StatusCreated, ""}, // 6
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 11, Title: "meat", TypeID: 1, Cost: 20}, http.StatusCreated, ""}, // 7
		{POST, SpendsPath, models.AddSpendReq{DayID: 11, Title: "egg", TypeID: 1, Cost: 7}, http.StatusCreated, ""},   // 8
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 12, Title: "pizza", TypeID: 3, Cost: 100}, http.StatusCreated, ""}, // 9
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 15, Title: "book American Gods", Notes: "as a gift", Cost: 30}, http.StatusCreated, ""}, // 10
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "new mirror in the bathroom", TypeID: 6, Cost: 150}, http.StatusCreated, ""}, // 11
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "new towels", TypeID: 6, Cost: 50}, http.StatusCreated, ""},                  // 12
		{POST, SpendsPath, models.AddSpendReq{DayID: 16, Title: "temp", Cost: 0}, http.StatusCreated, ""},                                    // 13
	} {
		var resp models.AddMonthlyPaymentResp
		req.Send(t, host, &resp)
		require.Equal(uint(i+1), resp.ID)
	}

	// Manage
	for _, req := range []Request{
		{PUT, SpendsPath, models.EditSpendReq{ID: 4, TypeID: ptrUint(0)}, http.StatusOK, ""},
		{PUT, SpendsPath, models.EditSpendReq{ID: 8, Title: ptrStr("eggs"), Notes: ptrStr("10 count"), Cost: ptrFloat(8)}, http.StatusOK, ""},
		{DELETE, SpendsPath, models.RemoveSpendReq{ID: 13}, http.StatusOK, ""},
	} {
		req.Send(t, host, nil)
	}

	// Check
	var resp models.GetMonthResp
	Request{GET, MonthsPath, models.GetMonthByIDReq{ID: 1}, http.StatusOK, ""}.Send(t, host, &resp)

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
	for i := len(expectedDays) + 1; i <= len(resp.Month.Days); i++ {
		expectedDays = append(expectedDays, db.Day{ID: uint(i), Spends: []db.Spend{}})
	}
	for i := range expectedDays {
		expectedDays[i].Year = resp.Month.Year
		expectedDays[i].Month = resp.Month.Month
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
		expectedDays[i].Saldo = prevSaldo.Add(resp.Month.DailyBudget)
		for _, s := range expectedDays[i].Spends {
			expectedDays[i].Saldo = expectedDays[i].Saldo.Sub(s.Cost)
		}
	}
	require.Equal(expectedDays, resp.Month.Days)

	checkMonth(require, 3050, -880, -894, resp.Month)
}
