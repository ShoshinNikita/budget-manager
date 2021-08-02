package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/web"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestBadRequests(t *testing.T) {
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
		{name: "get", f: testErrors_GetRequests},
		{name: "add", f: testErrors_AddRequests},
		{name: "edit", f: testErrors_EditRequests},
		{name: "remove", f: testErrors_RemoveRequests},
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

func testErrors_GetRequests(t *testing.T, host string) {
	for _, tt := range []struct {
		path Path
		req  interface{}
		err  string
		code int
	}{
		{SearchSpendsPath, models.SearchSpendsReq{MinCost: 10, MaxCost: 5}, "min_cost can't be greater than max_cost", http.StatusBadRequest},
		{MonthsPath, models.GetMonthByIDReq{ID: 10}, "such Month doesn't exist", http.StatusNotFound},
	} {
		Request{GET, tt.path, tt.req, tt.code, tt.err}.Send(t, host, nil)
	}
}

func testErrors_AddRequests(t *testing.T, host string) {
	for _, tt := range []struct {
		path Path
		req  interface{}
		err  string
	}{
		{IncomesPath, models.AddIncomeReq{MonthID: 0, Title: "1", Income: 1}, "month_id can't be empty or zero"},
		{IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "", Income: 1}, "title can't be empty"},
		{IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "", Income: 1}, "title can't be empty"},
		{IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "1", Income: 0}, "income must be greater than zero"},
		{IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "1", Income: -1}, "income must be greater than zero"},
		//
		{MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 0, Title: "1", Cost: 1}, "month_id can't be empty or zero"},
		{MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "	", Cost: 1}, "title can't be empty"},
		{MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "1", Cost: 0}, "cost must be greater than zero"},
		{MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "1", Cost: -1}, "cost must be greater than zero"},
		{MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "1", Cost: 1, TypeID: 10}, "such Spend Type doesn't exist"},
		//
		{SpendsPath, models.AddSpendReq{DayID: 0, Title: "1", Cost: 1}, "day_id can't be empty or zero"},
		{SpendsPath, models.AddSpendReq{DayID: 1, Title: "           ", Cost: 1}, "title can't be empty"},
		{SpendsPath, models.AddSpendReq{DayID: 1, Title: "1", Cost: -1}, "cost must be greater or equal to zero"},
		{SpendsPath, models.AddSpendReq{DayID: 1, Title: "1", Cost: 1, TypeID: 10}, "such Spend Type doesn't exist"},
		//
		{SpendTypesPath, models.AddSpendTypeReq{Name: "   "}, "name can't be empty"},
	} {
		Request{POST, tt.path, tt.req, http.StatusBadRequest, tt.err}.Send(t, host, nil)
	}
}

func testErrors_EditRequests(t *testing.T, host string) {
	// Not Found
	for _, tt := range []struct {
		path Path
		req  interface{}
		err  string
	}{
		{IncomesPath, models.EditIncomeReq{ID: 10, Title: ptrStr("new")}, "such Income doesn't exist"},
		{MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 10, Title: ptrStr("new")}, "such Monthly Payment doesn't exist"},
		{SpendsPath, models.EditSpendReq{ID: 10, Title: ptrStr("new")}, "such Spend doesn't exist"},
		{SpendTypesPath, models.EditSpendTypeReq{ID: 10, Name: ptrStr("new")}, "such Spend Type doesn't exist"},
	} {
		Request{PUT, tt.path, tt.req, http.StatusNotFound, tt.err}.Send(t, host, nil)
	}

	// Add entities to edit
	for _, req := range []RequestCreated{
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "title", Income: 100}},
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "title", Cost: 100}},
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "title", Cost: 100}},
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "#1"}},
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "#2", ParentID: 1}},
	} {
		req.Send(t, host, nil)
	}

	for _, tt := range []struct {
		path Path
		req  interface{}
		err  string
	}{
		{IncomesPath, models.EditIncomeReq{ID: 1, Title: ptrStr("")}, "title can't be empty"},
		{IncomesPath, models.EditIncomeReq{ID: 1, Title: ptrStr("          ")}, "title can't be empty"},
		{IncomesPath, models.EditIncomeReq{ID: 1, Income: ptrFloat(0)}, "income must be greater than zero"},
		{IncomesPath, models.EditIncomeReq{ID: 1, Income: ptrFloat(-11)}, "income must be greater than zero"},
		//
		{MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 1, Title: ptrStr("   ")}, "title can't be empty"},
		{MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 1, Cost: ptrFloat(0)}, "cost must be greater than zero"},
		{MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 1, Cost: ptrFloat(-11)}, "cost must be greater than zero"},
		{MonthlyPaymentsPath, models.EditMonthlyPaymentReq{ID: 1, TypeID: ptrUint(10)}, "such Spend Type doesn't exist"},
		//
		{SpendsPath, models.EditSpendReq{ID: 1, Title: ptrStr("	")}, "title can't be empty"},
		{SpendsPath, models.EditSpendReq{ID: 1, Cost: ptrFloat(-10)}, "cost must be greater or equal to zero"},
		{SpendsPath, models.EditSpendReq{ID: 1, TypeID: ptrUint(10)}, "such Spend Type doesn't exist"},
		//
		{SpendTypesPath, models.EditSpendTypeReq{ID: 1, Name: ptrStr("     ")}, "name can't be empty"},
		{SpendTypesPath, models.EditSpendTypeReq{ID: 1, ParentID: ptrUint(10)}, "check for a cycle failed: invalid Spend Type"},
		{SpendTypesPath, models.EditSpendTypeReq{ID: 1, ParentID: ptrUint(2)}, "Spend Type with new parent type will have a cycle"},
	} {
		Request{PUT, tt.path, tt.req, http.StatusBadRequest, tt.err}.Send(t, host, nil)
	}
}

func testErrors_RemoveRequests(t *testing.T, host string) {
	for _, tt := range []struct {
		path Path
		req  interface{}
		err  string
	}{
		{IncomesPath, models.RemoveIncomeReq{ID: 10}, "such Income doesn't exist"},
		{MonthlyPaymentsPath, models.RemoveMonthlyPaymentReq{ID: 10}, "such Monthly Payment doesn't exist"},
		{SpendsPath, models.RemoveSpendReq{ID: 10}, "such Spend doesn't exist"},
		{SpendTypesPath, models.RemoveSpendTypeReq{ID: 10}, "such Spend Type doesn't exist"},
	} {
		Request{DELETE, tt.path, tt.req, http.StatusNotFound, tt.err}.Send(t, host, nil)
	}
}
