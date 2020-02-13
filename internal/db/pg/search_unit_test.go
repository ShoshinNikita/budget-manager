package pg

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/stretchr/testify/require"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// nolint:lll
func TestBuildSearchSpendsQuery(t *testing.T) {
	t.Parallel()

	buildWhereQuery := func(whereQuery string) string {
		query := `
			SELECT spend.id AS id, month.year AS year, month.month AS month, day.day AS day,
			       spend.title AS title, spend.notes AS notes, spend.cost AS cost,
			       spend_type.id AS type__id, spend_type.name AS type__name

			 FROM "spends" AS "spend"
			      INNER JOIN days AS day
			      ON day.id = spend.day_id

			      INNER JOIN months AS month
			      ON month.id = day.month_id

			      LEFT JOIN spend_types AS spend_type
				  ON spend_type.id = spend.type_id`

		query += "\n" + whereQuery + "\n"

		query += `ORDER BY "month"."year", "month"."month", "day"."day", "spend"."id"`

		return formatQuery(query)
	}

	tests := []struct {
		desc     string
		args     db_common.SearchSpendsArgs
		reqQuery string
	}{
		// Notes:
		//  - we don't add ';' because orm.Query.AppendQuery returns query without it.
		//  - some names should be surrounded by quotes because go-pg escapes column and table names
		{
			desc:     "no args",
			reqQuery: buildWhereQuery(""), // empty WHERE clause
			args:     db_common.SearchSpendsArgs{},
		},
		{
			desc:     "specify title",
			reqQuery: buildWhereQuery(`WHERE (spend.title LIKE 'rent%')`),
			args: db_common.SearchSpendsArgs{
				Title: "rent%",
			},
		},
		{
			desc:     "specify title and notes",
			reqQuery: buildWhereQuery(`WHERE (spend.title LIKE 'rent%') AND (spend.notes LIKE '%note%')`),
			args: db_common.SearchSpendsArgs{
				Title: "rent%",
				Notes: "%note%",
			},
		},
		{
			desc: "specify after",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) >= '2018-01-15 15:37:00+00:00:00')`,
			),
			args: db_common.SearchSpendsArgs{
				After: time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
			},
		},
		{
			desc: "specify before",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) <= '2018-07-28 15:37:18+00:00:00')`,
			),
			args: db_common.SearchSpendsArgs{
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
		},
		{
			desc: "specify after and before",
			reqQuery: buildWhereQuery(
				`WHERE (make_date(month.year::int, month.month::int, day.day::int) BETWEEN 
				       '2018-01-15 15:37:00+00:00:00' AND '2018-07-28 15:37:18+00:00:00')`,
			),
			args: db_common.SearchSpendsArgs{
				After:  time.Date(2018, time.January, 15, 15, 37, 0, 0, time.UTC),
				Before: time.Date(2018, time.July, 28, 15, 37, 18, 0, time.UTC),
			},
		},
		{
			desc:     "specify min cost",
			reqQuery: buildWhereQuery(`WHERE (spend.cost >= 1535)`),
			args: db_common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
			},
		},
		{
			desc:     "specify max cost",
			reqQuery: buildWhereQuery(`WHERE (spend.cost <= 1500050)`),
			args: db_common.SearchSpendsArgs{
				MaxCost: money.FromFloat(15000.50),
			},
		},
		{
			desc:     "specify min and max costs",
			reqQuery: buildWhereQuery(`WHERE (spend.cost BETWEEN 1535 AND 1500050)`),
			args: db_common.SearchSpendsArgs{
				MinCost: money.FromFloat(15.35),
				MaxCost: money.FromFloat(15000.50),
			},
		},
		{
			desc:     "specify type ids",
			reqQuery: buildWhereQuery(`WHERE (spend.type_id IN (1,2,5,25,3))`),
			args: db_common.SearchSpendsArgs{
				TypeIDs: []uint{1, 2, 5, 25, 3},
			},
		},
		{
			desc: "all args",
			reqQuery: buildWhereQuery(`
				WHERE (spend.title LIKE '123')
				AND (spend.notes LIKE 'some note')
				AND (make_date(month.year::int, month.month::int, day.day::int)
					BETWEEN '2020-01-01 00:00:00+00:00:00' AND '2020-02-01 00:00:00+00:00:00')
				AND (spend.cost BETWEEN 20000 AND 500000)
				AND (spend.type_id IN (1,7))
			`),
			args: db_common.SearchSpendsArgs{
				Title:   "123",
				Notes:   "some note",
				After:   time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				Before:  time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
				MinCost: money.FromFloat(200),
				MaxCost: money.FromFloat(5000),
				TypeIDs: []uint{1, 7},
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
