// +build integration

package pg

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	. "github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// nolint:lll
func TestSearchSpends(t *testing.T) {
	globalRequire := require.New(t)

	db := initDB(globalRequire)
	defer cleanUp(globalRequire, db)

	// We have to drop and create tables for Months and Days because we manually prepare months for tests
	err := dropTables(db.db, &Month{}, nil, &Day{}, nil)
	globalRequire.Nil(err)
	err = createTables(db.db, &Month{}, nil, &Day{}, nil)
	globalRequire.Nil(err)

	// Preparations

	// Prepare months
	months := []struct {
		year  int
		month time.Month
	}{
		{year: 2019, month: time.December}, // 31 days (1 - 31 ids)
		{year: 2020, month: time.January},  // 31 days (32 - 62 ids)
		{year: 2020, month: time.February}, // 29 days (63 - 91 ids)
	}
	for _, m := range months {
		err := db.addMonth(m.year, m.month)
		globalRequire.Nil(err)
	}

	// Prepare Spend Types
	firstSpendType := &db_common.SpendType{ID: 1, Name: "first type"}
	secondSpendType := &db_common.SpendType{ID: 2, Name: "second type"}
	for _, t := range []*db_common.SpendType{firstSpendType, secondSpendType} {
		_, err := db.AddSpendType(context.Background(), t.Name)
		globalRequire.Nil(err)
	}

	// Prepare spends
	spends := []struct {
		db_common.Spend

		dayID uint
	}{
		//2019-12
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
			dayID: 1,
		},
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
			dayID: 5,
		},
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
			dayID: 8,
		},
		{
			Spend: db_common.Spend{Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
			dayID: 8,
		},
		// 2020-01
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
			dayID: 32,
		},
		{
			Spend: db_common.Spend{Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
			dayID: 32,
		},
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
			dayID: 41,
		},
		{
			Spend: db_common.Spend{Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
			dayID: 61,
		},
		// 2020-02
		{
			Spend: db_common.Spend{Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
			dayID: 70,
		},
		{
			Spend: db_common.Spend{Title: "qwerty", Notes: "2020-02-13", Cost: FromInt(7821)},
			dayID: 75,
		},
		{
			Spend: db_common.Spend{Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
			dayID: 76,
		},
	}
	for _, s := range spends {
		args := db_common.AddSpendArgs{DayID: s.dayID, Title: s.Title, Notes: s.Notes, Cost: s.Cost}
		if s.Type != nil {
			args.TypeID = s.Type.ID
		}

		_, err := db.AddSpend(context.Background(), args)
		globalRequire.Nil(err)
	}

	tests := []struct {
		desc string

		args db_common.SearchSpendsArgs
		want []*db_common.Spend
	}{
		// Tips: use this list of spends for creating future test cases
		{
			desc: "get all spends",
			args: db_common.SearchSpendsArgs{},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 6, Year: 2020, Month: time.January, Day: 1, Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 8, Year: 2020, Month: time.January, Day: 30, Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
				{ID: 10, Year: 2020, Month: time.February, Day: 13, Title: "qwerty", Notes: "2020-02-13", Cost: FromInt(7821)},
				{ID: 11, Year: 2020, Month: time.February, Day: 14, Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
			},
		},
		{
			desc: "get spends with 'first' in title",
			args: db_common.SearchSpendsArgs{
				Title: "first%",
			},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
				{ID: 11, Year: 2020, Month: time.February, Day: 14, Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
			},
		},
		{
			desc: "get spends with '2019-12-' in notes",
			args: db_common.SearchSpendsArgs{
				Notes: "2019-12-%",
			},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
			},
		},
		{
			desc: "get spends between '2019-12-06' and '2020-01-05'",
			args: db_common.SearchSpendsArgs{
				After:  time.Date(2019, time.December, 6, 0, 0, 0, 0, time.UTC),
				Before: time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
			},
			want: []*db_common.Spend{
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 6, Year: 2020, Month: time.January, Day: 1, Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
			},
		},
		{
			desc: "get spends with cost in range [13,100]",
			args: db_common.SearchSpendsArgs{
				MinCost: FromInt(13),
				MaxCost: FromInt(100),
			},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 8, Year: 2020, Month: time.January, Day: 30, Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
			},
		},
		{
			desc: "get spends with Spend Type with id 1",
			args: db_common.SearchSpendsArgs{
				TypeIDs: []uint{1},
			},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
			},
		},
		{
			desc: "get spends with Spend Type with id 1 and 2",
			args: db_common.SearchSpendsArgs{
				TypeIDs: []uint{1, 2},
			},
			want: []*db_common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
			},
		},
		{
			desc: "complex request",
			args: db_common.SearchSpendsArgs{
				Title:   "%spend",
				Notes:   "2020-%",
				After:   time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				Before:  time.Date(2020, time.February, 13, 0, 0, 0, 0, time.UTC),
				MinCost: FromInt(100),
				MaxCost: FromInt(7000),
			},
			want: []*db_common.Spend{
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			res, err := db.SearchSpends(context.Background(), tt.args)
			require.Nil(t, err)
			require.Equal(t, tt.want, res)
		})
	}
}
