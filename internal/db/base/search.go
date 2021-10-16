package base

import (
	"context"
	"strings"
	"time"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/types"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func (db DB) SearchSpends(ctx context.Context, args common.SearchSpendsArgs) ([]common.Spend, error) {
	var spends []struct {
		ID    uint         `db:"id"`
		Year  int          `db:"year"`
		Month time.Month   `db:"month"`
		Day   int          `db:"day"`
		Title string       `db:"title"`
		Notes types.String `db:"notes"`
		Cost  money.Money  `db:"cost"`

		Type SpendType `db:"type"`
	}
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		query, sqlArgs := db.buildSearchSpendsQuery(args)
		return tx.Select(&spends, query, sqlArgs...)
	})
	if err != nil {
		return nil, err
	}

	// Convert the internal model to the common one
	res := make([]common.Spend, 0, len(spends))
	for _, s := range spends {
		res = append(res, common.Spend{
			ID:    s.ID,
			Year:  s.Year,
			Month: s.Month,
			Day:   s.Day,
			Title: s.Title,
			Type:  s.Type.ToCommon(),
			Notes: string(s.Notes),
			Cost:  s.Cost,
		})
	}
	return res, nil
}

// buildSearchSpendsQuery builds a query to search for spends
//nolint:funlen
func (DB) buildSearchSpendsQuery(args common.SearchSpendsArgs) (string, []interface{}) {
	var (
		wheres    []string
		whereArgs []interface{}
	)
	addWhere := func(where string, args ...interface{}) {
		wheres = append(wheres, where)
		whereArgs = append(whereArgs, args...)
	}

	query := "SELECT "

	query += strings.Join([]string{
		`spend.id AS id`,
		`month.year AS year`,
		`month.month AS month`,
		`day.day AS day`,
		`spend.title AS title`,
		`spend.notes AS notes`,
		`spend.cost AS cost`,
		`spend_type.id AS "type.id"`,
		`spend_type.name AS "type.name"`,
		`spend_type.parent_id AS "type.parent_id"`,
	}, ", ")

	query += " FROM spends AS spend "

	query += strings.Join([]string{
		`INNER JOIN days AS day ON day.id = spend.day_id`,
		`INNER JOIN months AS month ON month.id = day.month_id`,
		`LEFT JOIN spend_types AS spend_type ON spend_type.id = spend.type_id`,
	}, " ")

	if args.Title != "" {
		title := "%" + args.Title + "%"
		if args.TitleExactly {
			title = args.Title
		}
		addWhere("LOWER(spend.title) LIKE ?", title)
	}

	if args.Notes != "" {
		notes := "%" + args.Notes + "%"
		if args.NotesExactly {
			notes = args.Notes
		}
		addWhere("LOWER(spend.notes) LIKE ?", notes)
	}

	if q, args := getQueryToFilterByTime(args.After, args.Before); q != "" {
		addWhere(q, args...)
	}

	switch {
	case args.MinCost != 0 && args.MaxCost != 0:
		addWhere("spend.cost BETWEEN ? AND ?", int(args.MinCost), int(args.MaxCost))
	case args.MinCost != 0:
		addWhere("spend.cost >= ?", int(args.MinCost))
	case args.MaxCost != 0:
		addWhere("spend.cost <= ?", int(args.MaxCost))
	}

	if len(args.TypeIDs) != 0 {
		var (
			orWheres    []string
			typeIDsArgs []interface{}
		)

		typeIDs := args.TypeIDs
		for i, id := range typeIDs {
			if id == 0 {
				// Search for spends without type
				orWheres = append(orWheres, "spend.type_id IS NULL")
				typeIDs = append(typeIDs[:i], typeIDs[i+1:]...)
				break
			}
		}

		if len(typeIDs) != 0 {
			inPlaceholders := strings.Repeat("?,", len(typeIDs))
			inPlaceholders = inPlaceholders[:len(inPlaceholders)-1]

			orWheres = append(orWheres, "spend.type_id IN ("+inPlaceholders+")")
			for _, id := range typeIDs {
				typeIDsArgs = append(typeIDsArgs, int(id))
			}
		}

		addWhere("("+strings.Join(orWheres, " OR ")+")", typeIDsArgs...)
	}

	var orders []string
	switch args.Sort {
	case common.SortSpendsByDate:
		orders = []string{"month.year", "month.month", "day.day"}
	case common.SortSpendsByTitle:
		orders = []string{"spend.title"}
	case common.SortSpendsByCost:
		orders = []string{"spend.cost"}
	}
	if args.Order == common.OrderByDesc {
		for i := range orders {
			orders[i] += " DESC"
		}
	}
	orders = append(orders, "spend.id")

	// Build the final query
	if len(wheres) != 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " ORDER BY " + strings.Join(orders, ", ")

	return query, whereArgs
}

func getQueryToFilterByTime(after, before time.Time) (where string, args []interface{}) {
	convertTime := func(t time.Time) int {
		return t.Year()*10000 + int(t.Month())*100 + t.Day()
	}

	// It is a db-agnostic solution to compare dates
	where = "month.year*10000 + month.month*100 + day.day"

	switch {
	case !after.IsZero() && !before.IsZero():
		where += " BETWEEN ? AND ?"
		args = []interface{}{convertTime(after), convertTime(before)}

	case !after.IsZero():
		where += " >= ?"
		args = []interface{}{convertTime(after)}

	case !before.IsZero():
		where += " <= ?"
		args = []interface{}{convertTime(before)}

	default:
		return "", nil
	}
	return where, args
}
