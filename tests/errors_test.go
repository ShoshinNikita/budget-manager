package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

func TestBadRequests(t *testing.T) {
	t.Parallel()

	RunTest(t, TestCases{
		{Name: "init", Fn: testErrors_InitData},
		{Name: "get", Fn: testErrors_GetRequests},
		{Name: "add", Fn: testErrors_AddRequests},
		{Name: "edit", Fn: testErrors_EditRequests},
		{Name: "remove", Fn: testErrors_RemoveRequests},
	})
}

func testErrors_InitData(t *testing.T, host string) {
	for _, req := range []RequestCreated{
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "#1"}},
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "#2", ParentID: 1}},
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "For monthly payments"}},
		{POST, SpendTypesPath, models.AddSpendTypeReq{Name: "For spends"}},
		//
		{POST, IncomesPath, models.AddIncomeReq{MonthID: 1, Title: "title", Income: 100}},
		//
		{POST, MonthlyPaymentsPath, models.AddMonthlyPaymentReq{MonthID: 1, Title: "title", TypeID: 3, Cost: 100}},
		//
		{POST, SpendsPath, models.AddSpendReq{DayID: 1, Title: "title", TypeID: 4, Cost: 100}},
	} {
		req.Send(t, host, nil)
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
		{MonthsPath, models.GetMonthByDateReq{Year: 2020, Month: time.January}, "such Month doesn't exist", http.StatusNotFound},
		{MonthsPath, models.GetMonthByDateReq{Year: 2020, Month: -1}, "invalid month", http.StatusBadRequest},
		{MonthsPath, models.GetMonthByDateReq{Year: 2020, Month: 0}, "invalid month", http.StatusBadRequest},
		{MonthsPath, models.GetMonthByDateReq{Year: 2020, Month: 13}, "invalid month", http.StatusBadRequest},
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

	// Bad Requests
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
		path   Path
		req    interface{}
		status int
		err    string
	}{
		// Not Found
		{IncomesPath, models.RemoveIncomeReq{ID: 10}, http.StatusNotFound, "such Income doesn't exist"},
		{MonthlyPaymentsPath, models.RemoveMonthlyPaymentReq{ID: 10}, http.StatusNotFound, "such Monthly Payment doesn't exist"},
		{SpendsPath, models.RemoveSpendReq{ID: 10}, http.StatusNotFound, "such Spend doesn't exist"},
		{SpendTypesPath, models.RemoveSpendTypeReq{ID: 10}, http.StatusNotFound, "such Spend Type doesn't exist"},
		// Bad Request
		{SpendTypesPath, models.RemoveSpendTypeReq{ID: 3}, http.StatusBadRequest, "Spend Type is used by Monthly Payment or Spend"},
		{SpendTypesPath, models.RemoveSpendTypeReq{ID: 4}, http.StatusBadRequest, "Spend Type is used by Monthly Payment or Spend"},
	} {
		Request{DELETE, tt.path, tt.req, tt.status, tt.err}.Send(t, host, nil)
	}
}
