package pg

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/stretchr/testify/require"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestBuildSearchSpendsQuery(t *testing.T) {
	t.Parallel()

	const defaultOrderByQuery = `ORDER BY "month"."year", "month"."month", "day"."day", "spend"."id"`

	buildWhereQuery := func(whereQuery string, orderByQuery string) string {
		query := `
			SELECT spend.id AS id, month.year AS year, month.month AS month, day.day AS day,
			       spend.title AS title, spend.notes AS notes, spend.cost AS cost,
			       spend_type.id AS type__id, spend_type.name AS type__name, spend_type.parent_id AS type__parent_id

			 FROM "spends" AS "spend"
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
		desc     string
		args     common.SearchSpendsArgs
		reqQuery string
	}{
		// Notes:
		//  - we don't add ';' because orm.Query.AppendQuery returns query without it.
		//  - some names should be surrounded by quotes because go-pg escapes column and table names
		{
			desc:     "no args",
			reqQuery: buildWhereQuery("", defaultOrderByQuery), // empty WHERE clause
			args:     common.SearchSpendsArgs{},
		},
		{
			desc:     "specify title",
			reqQuery: buildWhereQuery(`WHERE (LOWER(spend.title) LIKE '%rent%')`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Title: "rent",
			},
		},
		{
			desc:     "specify title (exactly)",
			reqQuery: buildWhereQuery(`WHERE (LOWER(spend.title) LIKE 'rent')`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Title:        "rent",
				TitleExactly: true,
			},
		},
		{
			desc:     "specify notes",
			reqQuery: buildWhereQuery(`WHERE (LOWER(spend.notes) LIKE '%note%')`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Notes: "note",
			},
		},
		{
			desc:     "specify notes (exactly)",
			reqQuery: buildWhereQuery(`WHERE (LOWER(spend.notes) LIKE 'note')`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Notes:        "note",
				NotesExactly: true,
			},
		},
		{
			desc: "specify title and notes",
			reqQuery: buildWhereQuery(
				`WHERE (LOWER(spend.title) LIKE 'rent') AND (LOWER(spend.notes) LIKE '%note%')`,
				defaultOrderByQuery,
			),
			args: common.SearchSpendsArgs{
				Title:        "rent",
				TitleExactly: true,
				Notes:        "note",
			},
		},
		{
			desc: "specify after",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) >= '2018-01-15 15:37:00+00:00:00')`,
				defaultOrderByQuery,
			),
			args: common.SearchSpendsArgs{
				After: time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
			},
		},
		{
			desc: "specify before",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) <= '2018-07-28 15:37:18+00:00:00')`,
				defaultOrderByQuery,
			),
			args: common.SearchSpendsArgs{
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
		},
		{
			desc: "specify after and before",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) BETWEEN
					   '2018-01-15 15:37:00+00:00:00' AND '2018-07-28 15:37:18+00:00:00')`,
				defaultOrderByQuery,
			),
			args: common.SearchSpendsArgs{
				After:  time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
		},
		{
			desc:     "specify min cost",
			reqQuery: buildWhereQuery(`WHERE (spend.cost >= 1535)`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
			},
		},
		{
			desc:     "specify max cost",
			reqQuery: buildWhereQuery(`WHERE (spend.cost <= 1500050)`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				MaxCost: money.FromFloat(15000.50),
			},
		},
		{
			desc:     "specify min and max costs",
			reqQuery: buildWhereQuery(`WHERE (spend.cost BETWEEN 1535 AND 1500050)`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
				MaxCost: money.FromFloat(15000.50),
			},
		},
		{
			desc:     "specify type ids",
			reqQuery: buildWhereQuery(`WHERE ((spend.type_id IN (1,2,5,25,3)))`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{1, 2, 5, 25, 3},
			},
		},
		{
			desc:     "without type",
			reqQuery: buildWhereQuery(`WHERE ((spend.type_id IS NULL))`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{0},
			},
		},
		{
			desc:     "with and without type",
			reqQuery: buildWhereQuery(`WHERE ((spend.type_id IS NULL) OR (spend.type_id IN (5,3)))`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				TypeIDs: []uint{5, 3, 0},
			},
		},
		{
			desc: "all args",
			reqQuery: buildWhereQuery(`
				WHERE (LOWER(spend.title) LIKE '%123%')
				      AND (LOWER(spend.notes) LIKE 'some note')
				      AND (make_date(month.year::int, month.month::int, day.day::int)
				          BETWEEN '2020-01-01 00:00:00+00:00:00' AND '2020-02-01 00:00:00+00:00:00')
				      AND (spend.cost BETWEEN 20000 AND 500000)
				      AND ((spend.type_id IS NULL) OR (spend.type_id IN (1,7)))
			`, defaultOrderByQuery),
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
		},
		{
			desc: "sort by date (desc)",
			reqQuery: buildWhereQuery(
				"", `ORDER BY "month"."year" DESC, "month"."month" DESC, "day"."day" DESC, "spend"."id"`,
			),
			args: common.SearchSpendsArgs{
				Order: common.OrderByDesc,
			},
		},
		{
			desc:     "sort by title",
			reqQuery: buildWhereQuery("", `ORDER BY "spend"."title", "spend"."id"`),
			args: common.SearchSpendsArgs{
				Sort: common.SortSpendsByTitle,
			},
		},
		{
			desc:     "sort by title (desc)",
			reqQuery: buildWhereQuery("", `ORDER BY "spend"."title" DESC, "spend"."id"`),
			args: common.SearchSpendsArgs{
				Sort:  common.SortSpendsByTitle,
				Order: common.OrderByDesc,
			},
		},
		{
			desc:     "sort by cost",
			reqQuery: buildWhereQuery("", `ORDER BY "spend"."cost", "spend"."id"`),
			args: common.SearchSpendsArgs{
				Sort: common.SortSpendsByCost,
			},
		},
		{
			desc:     "sql injection",
			reqQuery: buildWhereQuery(`WHERE (LOWER(spend.title) LIKE 'title''; OR 1=1--')`, defaultOrderByQuery),
			args: common.SearchSpendsArgs{
				Title:        "title'; OR 1=1--",
				TitleExactly: true,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			// We can use zero values DB and pg.Tx because buildSearchSpendsQuery just builds query
			// according to passed args
			searchQuery := (&DB{}).buildSearchSpendsQuery(&pg.Tx{}, tt.args)

			formatter := orm.NewFormatter()
			formatter.WithModel(searchQuery)
			query, err := searchQuery.AppendQuery(formatter, nil)
			require.Nil(t, err)
			require.Equal(t, tt.reqQuery, string(query))
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
