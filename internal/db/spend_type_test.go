package db

import (
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAddSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer func() {
		dropDB(db, require)
		db.Shutdown()
	}()

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
		id, err := db.AddSpendType(t.Name)
		if t.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
		require.Equal(t.ID, id)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		spendType, err := db.GetSpendType(t.ID)
		if t.isError {
			require.NotNil(pg.ErrNoRows, errors.Cause(err))
			continue
		}
		require.Nil(err)
		require.Equal(t.SpendType, *spendType)
	}

	var allSpendTypes []SpendType
	for _, t := range spendTypes {
		if t.isError {
			continue
		}

		allSpendTypes = append(allSpendTypes, t.SpendType)
	}

	dbSpendTypes, err := db.GetSpendTypes()
	require.Nil(err)
	require.ElementsMatch(allSpendTypes, dbSpendTypes)
}

func TestEditSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer func() {
		dropDB(db, require)
		db.Shutdown()
	}()

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
		id, err := db.AddSpendType(t.origin.Name)
		require.Nil(err)
		require.Equal(t.origin.ID, id)
	}

	// Edit Spend Types
	for _, t := range spendTypes {
		err := db.EditSpendType(t.edited.ID, t.edited.Name)
		if t.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		spendType, err := db.GetSpendType(t.origin.ID)
		if t.isError {
			require.Equal(t.origin, *spendType)
			continue
		}
		require.Nil(err)
		require.Equal(t.edited, *spendType)
	}
}

func TestDeleteSpendType(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer func() {
		dropDB(db, require)
		db.Shutdown()
	}()

	spendTypes := []SpendType{
		{ID: 1, Name: "first type"},
		{ID: 2, Name: "второй тип"},
	}

	// Add Spend Types
	for _, t := range spendTypes {
		id, err := db.AddSpendType(t.Name)
		require.Nil(err)
		require.Equal(t.ID, id)
	}

	// Delete all Spend Type
	for _, t := range spendTypes {
		db.RemoveSpendType(t.ID)
	}

	// Check Spend Types
	for _, t := range spendTypes {
		_, err := db.GetSpendType(t.ID)
		require.Equal(pg.ErrNoRows, errors.Cause(err))
	}

	// Try to delete Spend Type with invalid id
	err := db.RemoveSpendType(20)
	require.NotNil(err)
}
