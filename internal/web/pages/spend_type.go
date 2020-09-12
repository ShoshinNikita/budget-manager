package pages

import (
	"context"
	"fmt"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

type SpendType struct {
	db.SpendType
	// FullName is a composite name that contains names of parent Spend Types
	FullName string
}

func (h Handlers) getSpendTypesWithFullNames(ctx context.Context) ([]SpendType, error) {
	spendTypes, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		return nil, err
	}

	spendTypesMap := make(map[uint]db.SpendType, len(spendTypes))
	for _, t := range spendTypes {
		spendTypesMap[t.ID] = t
	}

	res := make([]SpendType, 0, len(spendTypes))
	for _, t := range spendTypes {
		res = append(res, SpendType{
			SpendType: t,
			FullName:  getSpendTypeFullName(spendTypesMap, t.ID),
		})
	}
	return res, nil
}

func getSpendTypeFullName(spendTypes map[uint]db.SpendType, typeID uint) string {
	const maxDepth = 15

	var getFullName func(currentDepth int, currentType db.SpendType) string
	getFullName = func(currentDepth int, currentType db.SpendType) string {
		if currentDepth >= maxDepth {
			return "..."
		}
		if currentType.ParentID == 0 {
			return currentType.Name
		}

		nextType := spendTypes[currentType.ParentID]
		if nextType.Name == "" {
			return currentType.Name
		}

		// Use thin spaces ' ' to separate names
		return fmt.Sprintf("%s / %s", getFullName(currentDepth+1, nextType), currentType.Name)
	}

	if spendType, ok := spendTypes[typeID]; ok {
		return getFullName(0, spendType)
	}
	return ""
}

// populateMonthlyPaymentsWithFullSpendTypeNames replaces Spend Type names to full ones
func populateMonthlyPaymentsWithFullSpendTypeNames(spendTypes []SpendType, monthlyPayments []db.MonthlyPayment) {
	fullNames := make(map[uint]string, len(spendTypes))
	for _, t := range spendTypes {
		fullNames[t.ID] = t.FullName
	}

	for i := range monthlyPayments {
		if monthlyPayments[i].Type != nil {
			if fullName, ok := fullNames[monthlyPayments[i].Type.ID]; ok {
				monthlyPayments[i].Type.Name = fullName
			}
		}
	}
}

// populateSpendsWithFullSpendTypeNames replaces Spend Type names to full ones
func populateSpendsWithFullSpendTypeNames(spendTypes []SpendType, spends []db.Spend) {
	fullNames := make(map[uint]string, len(spendTypes))
	for _, t := range spendTypes {
		fullNames[t.ID] = t.FullName
	}

	for i := range spends {
		if spends[i].Type != nil {
			if fullName, ok := fullNames[spends[i].Type.ID]; ok {
				spends[i].Type.Name = fullName
			}
		}
	}
}
