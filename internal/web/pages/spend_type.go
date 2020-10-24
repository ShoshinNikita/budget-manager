package pages

import (
	"fmt"
	"sort"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

type SpendType struct {
	db.SpendType

	// FullName is a composite name that contains names of parent Spend Types
	FullName string
	// parentSpendTypeIDs is a set of ids of parent Spend Types
	parentSpendTypeIDs map[uint]struct{}
}

// getSpendTypesWithFullNames returns sorted Spend Types with full name
func getSpendTypesWithFullNames(spendTypes []db.SpendType) []SpendType {
	spendTypesMap := make(map[uint]db.SpendType, len(spendTypes))
	for _, t := range spendTypes {
		spendTypesMap[t.ID] = t
	}

	res := make([]SpendType, 0, len(spendTypes))
	for _, t := range spendTypes {
		fullName, parentIDs := getSpendTypeFullName(spendTypesMap, t.ID)
		res = append(res, SpendType{
			SpendType: t,
			//
			FullName:           fullName,
			parentSpendTypeIDs: parentIDs,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].FullName < res[j].FullName
	})

	return res
}

func getSpendTypeFullName(spendTypes map[uint]db.SpendType, typeID uint) (name string, parentIDs map[uint]struct{}) {
	const maxDepth = 15

	parentIDs = make(map[uint]struct{})

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

		parentIDs[currentType.ParentID] = struct{}{}

		// Use thin spaces ' ' to separate names
		return fmt.Sprintf("%s / %s", getFullName(currentDepth+1, nextType), currentType.Name)
	}

	if spendType, ok := spendTypes[typeID]; ok {
		return getFullName(0, spendType), parentIDs
	}
	return "", nil
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
