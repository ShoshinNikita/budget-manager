package statistics

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestCalculateSpentBySpendType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		spendTypes []db.SpendType
		spends     []db.Spend
		//
		wantTypes map[uint]spendType
		wantDepth int
		//
		wantDatasets []SpentBySpendTypeDataset
	}{
		// No types
		{
			spendTypes: nil,
			spends: []db.Spend{
				{Cost: money.FromInt(100)},
				{Cost: money.FromInt(200)},
			},
			//
			wantDepth: 1,
			wantTypes: map[uint]spendType{
				0: {
					SpendType: db.SpendType{ID: 0, Name: "No Type"},
					Spent:     money.FromInt(300),
				},
			},
			//
			wantDatasets: []SpentBySpendTypeDataset{
				{{SpendTypeName: "No Type", Spent: money.FromInt(300)}},
			},
		},
		// No spends (all types will be ignored)
		{
			spendTypes: []db.SpendType{
				{ID: 1, Name: "1"},
				{ID: 2, Name: "2"},
				{ID: 11, Name: "1/1", ParentID: 1},
			},
			spends: nil,
			//
			wantDepth: 0,
			wantTypes: map[uint]spendType{},
			//
			wantDatasets: []SpentBySpendTypeDataset{},
		},
		{
			spendTypes: []db.SpendType{
				{ID: 1, Name: "1"},
				{ID: 11, Name: "1/1", ParentID: 1},
				{ID: 12, Name: "1/2", ParentID: 1},
				{ID: 123, Name: "1/2/3", ParentID: 12},
				{ID: 2, Name: "2"},
				{ID: 3, Name: "3"},
				{ID: 31, Name: "3/1", ParentID: 3},
				{ID: 32, Name: "3/2", ParentID: 3},
			},
			spends: []db.Spend{
				{Cost: money.FromInt(888)},
				{Cost: money.FromInt(100), Type: &db.SpendType{ID: 1}},
				{Cost: money.FromInt(1000), Type: &db.SpendType{ID: 11}},
				{Cost: money.FromInt(800), Type: &db.SpendType{ID: 123}},
				{Cost: money.FromInt(200), Type: &db.SpendType{ID: 2}},
				{Cost: money.FromInt(200), Type: &db.SpendType{ID: 31}},
				{Cost: money.FromInt(400), Type: &db.SpendType{ID: 32}},
			},
			//
			wantDepth: 3,
			wantTypes: map[uint]spendType{
				0: {
					SpendType: db.SpendType{ID: 0, Name: "No Type"},
					Spent:     money.FromInt(888),
				},
				1: {
					SpendType:   db.SpendType{ID: 1, Name: "1"},
					Spent:       money.FromInt(100 + 1000 + 800),
					childrenIDs: []uint{11, 12},
				},
				11: {
					SpendType: db.SpendType{ID: 11, Name: "1/1", ParentID: 1},
					Spent:     money.FromInt(1000),
				},
				12: {
					SpendType:   db.SpendType{ID: 12, Name: "1/2", ParentID: 1},
					Spent:       money.FromInt(800),
					childrenIDs: []uint{123},
				},
				123: {
					SpendType: db.SpendType{ID: 123, Name: "1/2/3", ParentID: 12},
					Spent:     money.FromInt(800),
				},
				2: {
					SpendType: db.SpendType{ID: 2, Name: "2"},
					Spent:     money.FromInt(200),
				},
				3: {
					SpendType:   db.SpendType{ID: 3, Name: "3"},
					Spent:       money.FromInt(600),
					childrenIDs: []uint{32, 31}, // children are sorted in descending order by field 'Spent'
				},
				31: {
					SpendType: db.SpendType{ID: 31, Name: "3/1", ParentID: 3},
					Spent:     money.FromInt(200),
				},
				32: {
					SpendType: db.SpendType{ID: 32, Name: "3/2", ParentID: 3},
					Spent:     money.FromInt(400),
				},
			},
			//
			wantDatasets: []SpentBySpendTypeDataset{
				{
					{SpendTypeName: "1", Spent: money.FromInt(1900)},

					{SpendTypeName: "No Type", Spent: money.FromInt(888)},

					{SpendTypeName: "3", Spent: money.FromInt(600)},

					{SpendTypeName: "2", Spent: money.FromInt(200)},
				},
				{
					{SpendTypeName: "1/1", Spent: money.FromInt(1000)},
					{SpendTypeName: "1/2", Spent: money.FromInt(800)},
					{SpendTypeName: "", Spent: money.FromInt(100)}, // left from 1

					{SpendTypeName: "", Spent: money.FromInt(888)}, // No Type

					{SpendTypeName: "3/2", Spent: money.FromInt(400)},
					{SpendTypeName: "3/1", Spent: money.FromInt(200)},

					{SpendTypeName: "", Spent: money.FromInt(200)}, // 2
				},
				{
					{SpendTypeName: "", Spent: money.FromInt(1000)}, // 1/1
					{SpendTypeName: "1/2/3", Spent: money.FromInt(800)},
					{SpendTypeName: "", Spent: money.FromInt(100)}, // left from 1

					{SpendTypeName: "", Spent: money.FromInt(888)}, // No Type

					{SpendTypeName: "", Spent: money.FromInt(400)}, // 3/2
					{SpendTypeName: "", Spent: money.FromInt(200)}, // 3/1

					{SpendTypeName: "", Spent: money.FromInt(200)}, // 2
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			gotTypes, gotDepth := prepareSpendTypesForDatasets(tt.spendTypes, tt.spends)
			require.Equal(t, tt.wantDepth, gotDepth)
			require.Equal(t, tt.wantTypes, gotTypes)

			gotDatasets := createSpentBySpendTypeDatasets(gotTypes, gotDepth)
			require.Equal(t, tt.wantDatasets, gotDatasets)
		})
	}
}
