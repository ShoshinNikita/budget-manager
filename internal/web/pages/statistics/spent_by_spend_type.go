package statistics

import (
	"sort"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type SpentBySpendTypeDataset []SpentBySpendTypeData

type SpentBySpendTypeData struct {
	SpendTypeName string      `json:"spend_type_name"`
	Spent         money.Money `json:"spent"`
}

func CalculateSpentBySpendType(spendTypes []db.SpendType, spends []db.Spend) []SpentBySpendTypeDataset {
	types, depth := prepareSpendTypesForDatasets(spendTypes, spends)

	return createSpentBySpendTypeDatasets(types, depth)
}

type spendType struct {
	db.SpendType

	// Spent is an amount of money spent by this type or its children
	Spent money.Money
	// childrenIDs is a list with ids of child Spend Types sorted in descending order by field 'Spent'
	childrenIDs []uint
}

//nolint:gofumpt
func prepareSpendTypesForDatasets(spendTypes []db.SpendType,
	spends []db.Spend) (types map[uint]spendType, maxChildDepth int) {

	// Init Spend Types. Use Spend Type with id 0 for Spends without a type
	types = make(map[uint]spendType, len(spendTypes)+1)
	types[0] = spendType{
		SpendType: db.SpendType{ID: 0, Name: "No Type"},
	}
	for _, t := range spendTypes {
		types[t.ID] = spendType{SpendType: t}
	}

	// Sum spend costs by Spend Type. If Spend Type has a parent, it also will be updated
	for _, spend := range spends {
		var typeID uint
		if spend.Type != nil {
			typeID = spend.Type.ID
		}

		t := types[typeID]
		t.Spent = t.Spent.Add(spend.Cost)
		types[typeID] = t
		for parentID := t.ParentID; parentID != 0; parentID = types[parentID].ParentID {
			parentType := types[parentID]
			parentType.Spent = parentType.Spent.Add(spend.Cost)
			types[parentID] = parentType
		}
	}

	// Filter types without Spends
	for id := range types {
		if types[id].Spent == 0 {
			delete(types, id)
		}
	}

	// Populate Spend Types with children ids and calculate max child depth
	for id := range types {
		var (
			depth = 1

			parentID = types[id].ParentID
			childID  = id
		)
		for parentID != 0 {
			depth++

			parentType := types[parentID]
			var found bool
			for _, id := range parentType.childrenIDs {
				if childID == id {
					found = true
					break
				}
			}
			if !found {
				parentType.childrenIDs = append(parentType.childrenIDs, childID)
			}
			types[parentID] = parentType

			parentID = parentType.ParentID
			childID = parentType.ID
		}

		if maxChildDepth < depth {
			maxChildDepth = depth
		}
	}
	for id := range types {
		t := types[id]
		sort.Slice(t.childrenIDs, func(i, j int) bool {
			return types[t.childrenIDs[i]].Spent > types[t.childrenIDs[j]].Spent
		})
	}

	return types, maxChildDepth
}

func createSpentBySpendTypeDatasets(types map[uint]spendType, depth int) []SpentBySpendTypeDataset {
	// Sort types in descending order to start dataset with the greatest values
	sortedTypes := make([]spendType, 0, len(types))
	for _, t := range types {
		sortedTypes = append(sortedTypes, t)
	}
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i].Spent > sortedTypes[j].Spent
	})

	// TODO: add type 'Other' to avoid small elements in datasets?

	datasets := make([]SpentBySpendTypeDataset, depth)
	for _, t := range sortedTypes {
		if t.ParentID != 0 {
			// Child Spend Types will be filled during processing of their parents
			continue
		}
		addSpendTypeToDatasets(types, datasets, t.ID, 0)
	}
	return datasets
}

//nolint:gofumpt
func addSpendTypeToDatasets(spendTypes map[uint]spendType, datasets []SpentBySpendTypeDataset,
	typeID uint, depth int) {

	spendType := spendTypes[typeID]
	datasets[depth] = append(datasets[depth], SpentBySpendTypeData{
		SpendTypeName: spendType.Name,
		Spent:         spendType.Spent,
	})

	// Fill next dataset level with child Spend Types (they are sorted in descending order)
	left := spendType.Spent
	for _, childID := range spendType.childrenIDs {
		child := spendTypes[childID]
		left = left.Sub(child.Spent)

		addSpendTypeToDatasets(spendTypes, datasets, childID, depth+1)
	}
	if left != 0 {
		// Fill all next dataset levels with left amount
		for i := depth + 1; i < len(datasets); i++ {
			datasets[i] = append(datasets[i], SpentBySpendTypeData{SpendTypeName: "", Spent: left})
		}
	}
}
