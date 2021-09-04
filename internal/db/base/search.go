package base

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func (db DB) SearchSpends(ctx context.Context, args common.SearchSpendsArgs) ([]common.Spend, error) {
	var spends []searchSpendsModel
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		return tx.SelectQuery(&spends, db.buildSearchSpendsQuery(args))
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
			Notes: s.Notes,
			Cost:  s.Cost,
		})
	}
	return res, nil
}

type searchSpendsModel struct {
	ID uint

	Year  int
	Month time.Month
	Day   int

	Title string
	Notes string
	Cost  money.Money
	Type  SpendType
}

// buildSearchSpendsQuery builds a query to search for spends
//
// Full query looks like:
//
//  SELECT spend.id AS id, month.year AS year, month.month AS month, day.day AS day,
//         spend.title AS title, spend.notes AS notes, spend.cost AS cost,
//         spend_type.id AS type__id, spend_type.name AS type__name
//
//    FROM spends AS spend
//         INNER JOIN days AS day
//         ON day.id = spend.day_id
//
//         INNER JOIN months AS month
//         ON month.id = day.month_id
//
//         LEFT JOIN spend_types AS spend_type
//         ON spend_type.id = spend.type_id
//
//   WHERE spend.title LIKE ':title_pattern'
//         AND spend.notes LIKE 'notes_pattern'
//         AND make_date(month.year::int, month.month::int, day.day::int) BETWEEN ':after'::date AND ':before'::date
//         AND spend.cost BETWEEN :min AND :max
//         AND spend.type_id IN (:type_ids)
//
//   ORDER BY month.year, month.month, day.day, spend.id;
//
//nolint:funlen
func (DB) buildSearchSpendsQuery(args common.SearchSpendsArgs) squirrel.SelectBuilder {
	query := squirrel.Select(
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
	).
		From("spends AS spend").
		//
		InnerJoin("days AS day ON day.id = spend.day_id").
		InnerJoin("months AS month ON month.id = day.month_id").
		LeftJoin("spend_types AS spend_type ON spend_type.id = spend.type_id")

	if args.Title != "" {
		title := "%" + args.Title + "%"
		if args.TitleExactly {
			title = args.Title
		}
		query = query.Where("LOWER(spend.title) LIKE ?", title)
	}

	if args.Notes != "" {
		notes := "%" + args.Notes + "%"
		if args.NotesExactly {
			notes = args.Notes
		}
		query = query.Where("LOWER(spend.notes) LIKE ?", notes)
	}

	switch {
	case !args.After.IsZero() && !args.Before.IsZero():
		query = query.Where(
			"make_date(month.year::int, month.month::int, day.day::int) BETWEEN ? AND ?",
			formatTime(args.After), formatTime(args.Before),
		)
	case !args.After.IsZero():
		query = query.Where("make_date(month.year::int, month.month::int, day.day::int) >= ?", formatTime(args.After))
	case !args.Before.IsZero():
		query = query.Where("make_date(month.year::int, month.month::int, day.day::int) <= ?", formatTime(args.Before))
	}

	switch {
	case args.MinCost != 0 && args.MaxCost != 0:
		query = query.Where("spend.cost BETWEEN ? AND ?", int(args.MinCost), int(args.MaxCost))
	case args.MinCost != 0:
		query = query.Where("spend.cost >= ?", int(args.MinCost))
	case args.MaxCost != 0:
		query = query.Where("spend.cost <= ?", int(args.MaxCost))
	}

	if len(args.TypeIDs) != 0 {
		query = query.Where(func() (or squirrel.Or) {
			// We have to convert []uint to []int because the limitation of github.com/Masterminds/squirrel:
			// https://github.com/Masterminds/squirrel/issues/114
			typeIDs := make([]int, 0, len(args.TypeIDs))
			for _, id := range args.TypeIDs {
				typeIDs = append(typeIDs, int(id))
			}

			for i, id := range typeIDs {
				if id == 0 {
					// Search for spends without type
					or = append(or, squirrel.Eq{"spend.type_id": nil})
					typeIDs = append(typeIDs[:i], typeIDs[i+1:]...)
					break
				}
			}

			if len(typeIDs) != 0 {
				or = append(or, squirrel.Eq{"spend.type_id": typeIDs})
			}
			return or
		}())
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
	query = query.OrderBy(orders...)

	return query
}

func formatTime(t time.Time) string {
	// TODO: should we use different layout? This layout is used by github.com/go-pg/pg
	const layout = "2006-01-02 15:04:05.999999999-07:00:00"
	return t.Format(layout)
}
