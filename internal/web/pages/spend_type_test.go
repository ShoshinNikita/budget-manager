package pages

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

func TestGetSpendTypeFullName(t *testing.T) {
	t.Parallel()

	spendTypes := map[uint]db.SpendType{
		1: {ID: 1, Name: "1", ParentID: 0},
		//
		2: {ID: 2, Name: "2", ParentID: 1},
		3: {ID: 3, Name: "3", ParentID: 1},
		4: {ID: 4, Name: "4", ParentID: 2},
		//
		5: {ID: 5, Name: "5", ParentID: 5},
		//
		6: {ID: 6, Name: "6", ParentID: 7},
		7: {ID: 7, Name: "7", ParentID: 6},
		//
		8: {ID: 8, Name: "8", ParentID: 9},
		9: {ID: 8, Name: "9", ParentID: 10},
	}

	joinFullName := func(names ...string) string {
		return strings.Join(names, " / ")
	}

	tests := []struct {
		typeID uint
		want   string
	}{
		{typeID: 1, want: joinFullName("1")},
		{typeID: 2, want: joinFullName("1", "2")},
		{typeID: 3, want: joinFullName("1", "3")},
		{typeID: 4, want: joinFullName("1", "2", "4")},
		{typeID: 5, want: joinFullName("...", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5", "5")},
		{typeID: 6, want: joinFullName("...", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6")},
		{typeID: 7, want: joinFullName("...", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7", "6", "7")},
		{typeID: 8, want: joinFullName("9", "8")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := getSpendTypeFullName(spendTypes, tt.typeID)
			require.Equal(t, tt.want, got)
		})
	}
}
