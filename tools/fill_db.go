// Add test data: Incomes, Monthly Payments, Spends and Spend Types.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	addIncomes()
	addSpendTypes()
	addMonthlyPayments()
	addSpends()
}

func addIncomes() {
	incomes := []map[string]interface{}{
		{
			"month_id": 1,
			"title":    "Salary",
			"notes":    "short notes",
			"income":   65000,
		},
		{
			"month_id": 1,
			"title":    "Salary",
			"income":   180000,
		},
		{
			"month_id": 1,
			"title":    "Birthday Gift",
			"notes":    "some very very long notes with a lot of words",
			"income":   1000,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"income":   1875,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"income":   2000,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"income":   4546,
		},
	}

	fmt.Println("Add Incomes...")
	for i, in := range incomes {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(in) //nolint:errcheck
		resp, _ := http.Post("http://localhost:8080/api/incomes", "application/json", body)
		fmt.Printf("  [%d] code: %d\n", i+1, resp.StatusCode)
	}
}

func addSpendTypes() {
	types := []map[string]interface{}{
		{"name": "Rent"},
		{"name": "Internet"},
		{"name": "Food"},
	}

	fmt.Println("Add Spend Types...")
	for i, in := range types {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(in) //nolint:errcheck
		resp, _ := http.Post("http://localhost:8080/api/spend-types", "application/json", body)
		fmt.Printf("  [%d] code: %d\n", i+1, resp.StatusCode)
	}
}

func addMonthlyPayments() {
	incomes := []map[string]interface{}{
		{
			"month_id": 1,
			"title":    "Rent",
			"type_id":  1,
			"notes":    "short notes",
			"cost":     30000,
		},
		{
			"month_id": 1,
			"title":    "Patreon",
			"type_id":  2,
			"cost":     1000,
		},
		{
			"month_id": 1,
			"title":    "Netflix",
			"type_id":  2,
			"notes":    "some very very long notes with a lot of words",
			"cost":     2000,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"cost":     1875,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"cost":     2000,
		},
		{
			"month_id": 1,
			"title":    "123456",
			"cost":     4546,
		},
	}

	fmt.Println("Add Incomes...")
	for i, in := range incomes {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(in) //nolint:errcheck
		resp, _ := http.Post("http://localhost:8080/api/monthly-payments", "application/json", body)
		fmt.Printf("  [%d] code: %d\n", i+1, resp.StatusCode)
	}
}

func addSpends() {
	spends := []map[string]interface{}{
		// First day
		{
			"day_id":  1,
			"title":   "Bread",
			"type_id": 3,
			"notes":   "short notes",
			"cost":    50,
		},
		{
			"day_id":  1,
			"title":   "Butter",
			"type_id": 3,
			"cost":    300,
		},
		{
			"day_id":  1,
			"title":   "Sausage",
			"type_id": 3,
			"notes":   "some very very long notes with a lot of words",
			"cost":    1000,
		},
		// Fifth day
		{
			"day_id": 5,
			"title":  "Washing powder",
			"cost":   500,
		},
		{
			"day_id": 5,
			"title":  "Napkin",
			"notes":  "100x",
			"cost":   125.90,
		},
	}

	fmt.Println("Add Spends...")
	for i, in := range spends {
		body := bytes.NewBuffer(nil)
		json.NewEncoder(body).Encode(in) //nolint:errcheck
		resp, _ := http.Post("http://localhost:8080/api/spends", "application/json", body)
		fmt.Printf("  [%d] code: %d\n", i+1, resp.StatusCode)
	}
}
