// +build integration

package pg

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	. "github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

//nolint:lll
func TestSearchSpends(t *testing.T) {
	globalRequire := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	// Reset tables 'months' and 'days' because we manually prepare months for tests
	_, err := db.db.Exec(`DELETE FROM days; DELETE FROM months;`)
	globalRequire.Nil(err)
	// Reset sequences
	_, err = db.db.Exec(`ALTER SEQUENCE days_id_seq RESTART; ALTER SEQUENCE months_id_seq RESTART;`)
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
		err := db.initMonth(context.Background(), m.year, m.month)
		globalRequire.Nil(err)
	}

	// Prepare Spend Types
	firstSpendType := &common.SpendType{ID: 1, Name: "first type"}
	secondSpendType := &common.SpendType{ID: 2, Name: "second type"}
	for _, t := range []common.SpendType{*firstSpendType, *secondSpendType} {
		_, err := db.AddSpendType(context.Background(), common.AddSpendTypeArgs{Name: t.Name})
		globalRequire.Nil(err)
	}

	// Prepare spends
	spends := []struct {
		common.Spend

		dayID uint
	}{
		// 2019-12
		{
			Spend: common.Spend{Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
			dayID: 1,
		},
		{
			Spend: common.Spend{Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
			dayID: 5,
		},
		{
			Spend: common.Spend{Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
			dayID: 8,
		},
		{
			Spend: common.Spend{Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
			dayID: 8,
		},
		// 2020-01
		{
			Spend: common.Spend{Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
			dayID: 32,
		},
		{
			Spend: common.Spend{Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
			dayID: 32,
		},
		{
			Spend: common.Spend{Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
			dayID: 41,
		},
		{
			Spend: common.Spend{Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
			dayID: 61,
		},
		// 2020-02
		{
			Spend: common.Spend{Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
			dayID: 70,
		},
		{
			Spend: common.Spend{Title: "qwerty", Notes: "2020-02-13", Cost: FromInt(7821)},
			dayID: 75,
		},
		{
			Spend: common.Spend{Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
			dayID: 76,
		},
		{
			Spend: common.Spend{Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			dayID: 77,
		},
	}
	for _, s := range spends {
		args := common.AddSpendArgs{DayID: s.dayID, Title: s.Title, Notes: s.Notes, Cost: s.Cost}
		if s.Type != nil {
			args.TypeID = s.Type.ID
		}

		_, err := db.AddSpend(context.Background(), args)
		globalRequire.Nil(err)
	}

	tests := []struct {
		desc string

		args common.SearchSpendsArgs
		want []common.Spend
	}{
		// Tips: use this list of spends for creating future test cases
		{
			desc: "get all spends",
			args: common.SearchSpendsArgs{},
			want: []common.Spend{
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
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			},
		},
		{
			desc: "get spends with 'first spend' in title",
			args: common.SearchSpendsArgs{
				Title: "first spend",
			},
			want: []common.Spend{
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
			desc: "get spends with 'first spend' in title (exactly)",
			args: common.SearchSpendsArgs{
				Title:        "first spend",
				TitleExactly: true,
			},
			want: []common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
			},
		},
		{
			desc: "get spends with '2019-12-' in notes",
			args: common.SearchSpendsArgs{
				Notes: "2019-12-",
			},
			want: []common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
			},
		},
		{
			desc: "get spends between '2019-12-06' and '2020-01-05'",
			args: common.SearchSpendsArgs{
				After:  time.Date(2019, time.December, 6, 0, 0, 0, 0, time.UTC),
				Before: time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
			},
			want: []common.Spend{
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 6, Year: 2020, Month: time.January, Day: 1, Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
			},
		},
		{
			desc: "get spends with cost in range [13,100]",
			args: common.SearchSpendsArgs{
				MinCost: FromInt(13),
				MaxCost: FromInt(100),
			},
			want: []common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 8, Year: 2020, Month: time.January, Day: 30, Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
			},
		},
		{
			desc: "get spends with Spend Type with id 1",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{1},
			},
			want: []common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
			},
		},
		{
			desc: "get spends with Spend Type with id 1 and 2",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{1, 2},
			},
			want: []common.Spend{
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
			},
		},
		{
			desc: "complex request",
			args: common.SearchSpendsArgs{
				Title:   "spend",
				Notes:   "2020-",
				After:   time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
				Before:  time.Date(2020, time.February, 13, 0, 0, 0, 0, time.UTC),
				MinCost: FromInt(100),
				MaxCost: FromInt(7000),
			},
			want: []common.Spend{
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
			},
		},
		{
			desc: "search uppercase title with lowerace",
			args: common.SearchSpendsArgs{
				Title: "title",
			},
			want: []common.Spend{
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			},
		},
		{
			desc: "search uppercase title with uppercase (no rows)",
			args: common.SearchSpendsArgs{
				Title: "TITLE",
			},
			want: []common.Spend{},
		},
		{
			desc: "search uppercase notes with lowerace",
			args: common.SearchSpendsArgs{
				Notes: "notes",
			},
			want: []common.Spend{
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			},
		},
		{
			desc: "search uppercase notes with uppercase (no rows)",
			args: common.SearchSpendsArgs{
				Notes: "NOTES",
			},
			want: []common.Spend{},
		},
		{
			desc: "sql injection",
			args: common.SearchSpendsArgs{
				Title:        "first spend' OR 1=1--",
				TitleExactly: true,
			},
			want: []common.Spend{},
		},
		{
			desc: "search for spends without type",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{0},
			},
			want: []common.Spend{
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 6, Year: 2020, Month: time.January, Day: 1, Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 8, Year: 2020, Month: time.January, Day: 30, Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
				{ID: 10, Year: 2020, Month: time.February, Day: 13, Title: "qwerty", Notes: "2020-02-13", Cost: FromInt(7821)},
				{ID: 11, Year: 2020, Month: time.February, Day: 14, Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			},
		},
		{
			desc: "search for spends with and without type",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{0, 1, 2},
			},
			want: []common.Spend{
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
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
			},
		},
		{
			desc: "sort by cost, desc",
			args: common.SearchSpendsArgs{
				Sort:  common.SortSpendsByCost,
				Order: common.OrderByDesc,
			},
			want: []common.Spend{
				{ID: 6, Year: 2020, Month: time.January, Day: 1, Title: "second spend", Notes: "2020-01-01", Cost: FromInt(7821)},
				{ID: 10, Year: 2020, Month: time.February, Day: 13, Title: "qwerty", Notes: "2020-02-13", Cost: FromInt(7821)},
				{ID: 7, Year: 2020, Month: time.January, Day: 10, Title: "first spend", Notes: "2020-01-10", Cost: FromInt(555)},
				{ID: 11, Year: 2020, Month: time.February, Day: 14, Title: "first spending", Notes: "2020-02-14", Cost: FromInt(555)},
				{ID: 5, Year: 2020, Month: time.January, Day: 1, Title: "first spend", Notes: "2020-01-01", Cost: FromInt(189), Type: secondSpendType},
				{ID: 9, Year: 2020, Month: time.February, Day: 8, Title: "first spend", Notes: "2020-02-08", Cost: FromInt(189)},
				{ID: 3, Year: 2019, Month: time.December, Day: 8, Title: "first spend", Notes: "2019-12-08", Cost: FromInt(159)},
				{ID: 1, Year: 2019, Month: time.December, Day: 1, Title: "first spend", Notes: "2019-12-01", Cost: FromInt(100), Type: firstSpendType},
				{ID: 4, Year: 2019, Month: time.December, Day: 8, Title: "second spend", Notes: "2019-12-08", Cost: FromInt(15)},
				{ID: 8, Year: 2020, Month: time.January, Day: 30, Title: "second spend", Notes: "2020-01-30", Cost: FromInt(15)},
				{ID: 2, Year: 2019, Month: time.December, Day: 5, Title: "first spend", Notes: "2019-12-05", Cost: FromInt(10)},
				{ID: 12, Year: 2020, Month: time.February, Day: 15, Title: "TITLE", Notes: "NOTES", Cost: FromInt(1)},
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
