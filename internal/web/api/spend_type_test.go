package api

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

func TestCheckSpendTypeForCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		//
		spendTypes  []db.SpendType
		originalID  uint
		newParentID uint
		//
		wantErr      string
		wantHasCycle bool
	}{
		{
			desc: "no cycle",
			//
			spendTypes: []db.SpendType{
				{ID: 1},
				{ID: 2, ParentID: 3},
				{ID: 3},
			},
			originalID:  1,
			newParentID: 2,
			//
			wantErr:      "",
			wantHasCycle: false,
		},
		{
			desc: "has cycle",
			//
			spendTypes: []db.SpendType{
				{ID: 1},
				{ID: 2, ParentID: 3},
				{ID: 3, ParentID: 1},
			},
			originalID:  1,
			newParentID: 2,
			//
			wantErr:      "",
			wantHasCycle: true,
		},
		{
			desc: "already has cycle",
			//
			spendTypes: []db.SpendType{
				{ID: 1},
				{ID: 2, ParentID: 3},
				{ID: 3, ParentID: 2},
			},
			originalID:  1,
			newParentID: 2,
			//
			wantErr:      "Spend Type has too many parents or already has a cycle",
			wantHasCycle: false,
		},
		{
			desc: "invalid Spend Type",
			//
			spendTypes: []db.SpendType{
				{ID: 1},
				{ID: 2, ParentID: 4},
			},
			originalID:  1,
			newParentID: 2,
			//
			wantErr:      "invalid Spend Type",
			wantHasCycle: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			gotHasCycle, gotErr := checkSpendTypeForCycle(tt.spendTypes, tt.originalID, tt.newParentID)
			if tt.wantErr != "" {
				require.EqualError(t, gotErr, tt.wantErr)
			} else {
				require.Nil(t, gotErr)
			}
			require.Equal(t, tt.wantHasCycle, gotHasCycle)
		})
	}
}
