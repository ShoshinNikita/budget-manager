package base

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestBuildSearchSpendsQuery(t *testing.T) {
	t.Parallel()

	const defaultOrderByQuery = `ORDER BY month.year, month.month, day.day, spend.id`

	buildWhereQuery := func(whereQuery string, orderByQuery string) string {
		query := `
			SELECT spend.id AS id, month.year AS year, month.month AS month, day.day AS day,
			       spend.title AS title, spend.notes AS notes, spend.cost AS cost,
			       spend_type.id AS "type.id", spend_type.name AS "type.name", spend_type.parent_id AS "type.parent_id"

			 FROM spends AS spend
			      INNER JOIN days AS day
			      ON day.id = spend.day_id

			      INNER JOIN months AS month
			      ON month.id = day.month_id

			      LEFT JOIN spend_types AS spend_type
				  ON spend_type.id = spend.type_id`

		query += "\n" + whereQuery + "\n"

		query += orderByQuery

		return formatQuery(query)
	}

	tests := []struct {
		desc      string
		args      common.SearchSpendsArgs
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			desc:      "no args",
			wantQuery: buildWhereQuery("", defaultOrderByQuery), // empty WHERE clause
			args:      common.SearchSpendsArgs{},
		},
		{
			desc: "specify title",
			args: common.SearchSpendsArgs{
				Title: "rent",
			},
			wantQuery: buildWhereQuery(`WHERE LOWER(spend.title) LIKE ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{"%rent%"},
		},
		{
			desc: "specify title (exactly)",
			args: common.SearchSpendsArgs{
				Title:        "rent",
				TitleExactly: true,
			},
			wantQuery: buildWhereQuery(`WHERE LOWER(spend.title) LIKE ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{"rent"},
		},
		{
			desc: "specify notes",
			args: common.SearchSpendsArgs{
				Notes: "note",
			},
			wantQuery: buildWhereQuery(`WHERE LOWER(spend.notes) LIKE ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{"%note%"},
		},
		{
			desc: "specify notes (exactly)",
			args: common.SearchSpendsArgs{
				Notes:        "note",
				NotesExactly: true,
			},
			wantQuery: buildWhereQuery(`WHERE LOWER(spend.notes) LIKE ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{"note"},
		},
		{
			desc: "specify title and notes",
			args: common.SearchSpendsArgs{
				Title:        "rent",
				TitleExactly: true,
				Notes:        "note",
			},
			wantQuery: buildWhereQuery(
				`WHERE LOWER(spend.title) LIKE ? AND LOWER(spend.notes) LIKE ?`,
				defaultOrderByQuery,
			),
			wantArgs: []interface{}{"rent", "%note%"},
		},
		{
			desc: "specify after",
			args: common.SearchSpendsArgs{
				After: time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
			},
			wantQuery: buildWhereQuery(
				`WHERE month.year*10000 + month.month*100 + day.day >= ?`,
				defaultOrderByQuery,
			),
			wantArgs: []interface{}{20180115},
		},
		{
			desc: "specify before",
			args: common.SearchSpendsArgs{
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
			wantQuery: buildWhereQuery(
				`WHERE month.year*10000 + month.month*100 + day.day <= ?`,
				defaultOrderByQuery,
			),
			wantArgs: []interface{}{20180728},
		},
		{
			desc: "specify after and before",
			args: common.SearchSpendsArgs{
				After:  time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
			wantQuery: buildWhereQuery(
				`WHERE month.year*10000 + month.month*100 + day.day BETWEEN ? AND ?`,
				defaultOrderByQuery,
			),
			wantArgs: []interface{}{20180115, 20180728},
		},
		{
			desc: "specify min cost",
			args: common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
			},
			wantQuery: buildWhereQuery(`WHERE spend.cost >= ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{1535},
		},
		{
			desc: "specify max cost",
			args: common.SearchSpendsArgs{
				MaxCost: money.FromFloat(15000.50),
			},
			wantQuery: buildWhereQuery(`WHERE spend.cost <= ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{1500050},
		},
		{
			desc: "specify min and max costs",
			args: common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
				MaxCost: money.FromFloat(15000.50),
			},
			wantQuery: buildWhereQuery(`WHERE spend.cost BETWEEN ? AND ?`, defaultOrderByQuery),
			wantArgs:  []interface{}{1535, 1500050},
		},
		{
			desc: "specify type ids",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{1, 2, 5, 25, 3},
			},
			wantQuery: buildWhereQuery(`WHERE (spend.type_id IN (?,?,?,?,?))`, defaultOrderByQuery),
			wantArgs:  []interface{}{1, 2, 5, 25, 3},
		},
		{
			desc: "without type",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{0},
			},
			wantQuery: buildWhereQuery(`WHERE (spend.type_id IS NULL)`, defaultOrderByQuery),
		},
		{
			desc: "with and without type",
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{5, 3, 0},
			},
			wantQuery: buildWhereQuery(`WHERE (spend.type_id IS NULL OR spend.type_id IN (?,?))`, defaultOrderByQuery),
			wantArgs:  []interface{}{5, 3},
		},
		{
			desc: "all args",
			args: common.SearchSpendsArgs{
				Title:        "123",
				Notes:        "some note",
				NotesExactly: true,
				After:        time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				Before:       time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
				MinCost:      money.FromFloat(200),
				MaxCost:      money.FromFloat(5000),
				TypeIDs:      []uint{0, 1, 7},
			},
			wantQuery: buildWhereQuery(`
				WHERE LOWER(spend.title) LIKE ?
					  AND LOWER(spend.notes) LIKE ?
					  AND month.year*10000 + month.month*100 + day.day BETWEEN ? AND ?
					  AND spend.cost BETWEEN ? AND ?
					  AND (spend.type_id IS NULL OR spend.type_id IN (?,?))
			`, defaultOrderByQuery),
			wantArgs: []interface{}{
				"%123%",
				"some note",
				20200101, 20200201,
				20000, 500000,
				1, 7,
			},
		},
		{
			desc: "sort by date (desc)",
			args: common.SearchSpendsArgs{
				Order: common.OrderByDesc,
			},
			wantQuery: buildWhereQuery(
				"", `ORDER BY month.year DESC, month.month DESC, day.day DESC, spend.id`,
			),
		},
		{
			desc: "sort by title",
			args: common.SearchSpendsArgs{
				Sort: common.SortSpendsByTitle,
			},
			wantQuery: buildWhereQuery("", `ORDER BY spend.title, spend.id`),
		},
		{
			desc: "sort by title (desc)",
			args: common.SearchSpendsArgs{
				Sort:  common.SortSpendsByTitle,
				Order: common.OrderByDesc,
			},
			wantQuery: buildWhereQuery("", `ORDER BY spend.title DESC, spend.id`),
		},
		{
			desc: "sort by cost",
			args: common.SearchSpendsArgs{
				Sort: common.SortSpendsByCost,
			},
			wantQuery: buildWhereQuery("", `ORDER BY spend.cost, spend.id`),
		},
		{
			desc:      "sql injection",
			wantQuery: buildWhereQuery(`WHERE LOWER(spend.title) LIKE ?`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Title:        "title'; OR 1=1--",
				TitleExactly: true,
			},
			wantArgs: []interface{}{"title'; OR 1=1--"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			// We can use zero values DB and pg.Tx because buildSearchSpendsQuery just builds query
			// according to passed args
			searchQuery := (&DB{}).buildSearchSpendsQuery(tt.args)
			query, args, err := searchQuery.ToSql()
			require.Nil(t, err)
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

func formatQuery(query string) string {
	queryBuilder := strings.Builder{}

	buff := bytes.NewBuffer([]byte(query))
	scanner := bufio.NewScanner(buff)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Trim tabs, spaces and new line
		line = bytes.TrimSpace(line)

		if len(line) != 0 {
			queryBuilder.Write(line)
			queryBuilder.WriteByte(' ')
		}
	}
	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	formattedQuery := queryBuilder.String()
	formattedQuery = formattedQuery[:len(formattedQuery)-1] // trim the trailing space
	return formattedQuery
}
