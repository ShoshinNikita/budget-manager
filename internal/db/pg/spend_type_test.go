// +build integration

package pg

import (
	"context"
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

func TestAddSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	spendTypes := []struct {
		SpendType
		isError bool
	}{
		{
			SpendType: SpendType{ID: 1, Name: "first type"},
		},
		{
			SpendType: SpendType{ID: 2, Name: "второй тип"},
		},
		{
			SpendType: SpendType{ID: 0, Name: ""},
			isError:   true,
		},
	}

	// Add Spend Types
	for _, t := range spendTypes {
		id, err := db.AddSpendType(context.Background(), t.Name)
		if t.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
		require.Equal(t.ID, id)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		spendType, err := db.GetSpendType(context.Background(), t.ID)
		if t.isError {
			require.NotNil(pg.ErrNoRows, errors.Cause(err))
			continue
		}
		require.Nil(err)
		require.Equal(t.SpendType.ToCommon(), spendType)
	}

	allSpendTypes := make([]*db_common.SpendType, 0, len(spendTypes))
	for _, t := range spendTypes {
		if t.isError {
			continue
		}

		allSpendTypes = append(allSpendTypes, t.SpendType.ToCommon())
	}

	dbSpendTypes, err := db.GetSpendTypes(context.Background())
	require.Nil(err)
	require.ElementsMatch(allSpendTypes, dbSpendTypes)
}

func TestEditSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	spendTypes := []struct {
		origin  SpendType
		edited  SpendType
		isError bool
	}{
		{
			origin: SpendType{ID: 1, Name: "first type"},
			edited: SpendType{ID: 1, Name: "new name"},
		},
		{
			origin:  SpendType{ID: 2, Name: "first type"},
			edited:  SpendType{ID: 2, Name: ""},
			isError: true,
		},
	}

	// Add Spend Types
	for _, t := range spendTypes {
		id, err := db.AddSpendType(context.Background(), t.origin.Name)
		require.Nil(err)
		require.Equal(t.origin.ID, id)
	}

	// Edit Spend Types
	for _, t := range spendTypes {
		err := db.EditSpendType(context.Background(), t.edited.ID, t.edited.Name)
		if t.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		spendType, err := db.GetSpendType(context.Background(), t.origin.ID)
		if t.isError {
			require.Equal(t.origin.ToCommon(), spendType)
			continue
		}
		require.Nil(err)
		require.Equal(t.edited.ToCommon(), spendType)
	}
}

func TestDeleteSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	spendTypes := []SpendType{
		{ID: 1, Name: "first type"},
		{ID: 2, Name: "второй тип"},
	}

	// Add Spend Types
	for _, t := range spendTypes {
		id, err := db.AddSpendType(context.Background(), t.Name)
		require.Nil(err)
		require.Equal(t.ID, id)
	}

	// Delete all Spend Type
	for _, t := range spendTypes {
		err := db.RemoveSpendType(context.Background(), t.ID)
		require.Nil(err)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		_, err := db.GetSpendType(context.Background(), t.ID)
		require.Equal("such Spend Type doesn't exist", err.Error())
	}

	// Try to delete Spend Type with invalid id
	err := db.RemoveSpendType(context.Background(), 20)
	require.NotNil(err)
}
