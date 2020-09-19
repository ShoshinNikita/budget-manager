package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func (db DB) SearchSpends(ctx context.Context, args common.SearchSpendsArgs) ([]common.Spend, error) {
	var pgSpends []searchSpendsModel
	err := db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return db.buildSearchSpendsQuery(tx, args).Select(&pgSpends)
	})
	if err != nil {
		return nil, err
	}

	// Convert the internal model to the common one
	spends := make([]common.Spend, 0, len(pgSpends))
	for i := range pgSpends {
		s := common.Spend{
			ID:    pgSpends[i].ID,
			Year:  pgSpends[i].Year,
			Month: pgSpends[i].Month,
			Day:   pgSpends[i].Day,
			Title: pgSpends[i].Title,
			Notes: pgSpends[i].Notes,
			Cost:  pgSpends[i].Cost,
		}
		if pgSpends[i].Type.ID != 0 {
			// Don't check if a name is empty because Spend Types with non-zero id always have a name
			s.Type = pgSpends[i].Type.ToCommon()
		}
		spends = append(spends, s)
	}

	return spends, nil
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
func (DB) buildSearchSpendsQuery(tx *pg.Tx, args common.SearchSpendsArgs) *orm.Query {
	query := tx.ModelContext(tx.Context(), (*Spend)(nil)).
		ColumnExpr("spend.id AS id").
		ColumnExpr("month.year AS year").
		ColumnExpr("month.month AS month").
		ColumnExpr("day.day AS day").
		ColumnExpr("spend.title AS title").
		ColumnExpr("spend.notes AS notes").
		ColumnExpr("spend.cost AS cost").
		ColumnExpr("spend_type.id AS type__id").
		ColumnExpr("spend_type.name AS type__name").
		ColumnExpr("spend_type.parent_id AS type__parent_id").
		//
		Join("INNER JOIN days AS day ON day.id = spend.day_id").
		Join("INNER JOIN months AS month ON month.id = day.month_id").
		Join("LEFT JOIN spend_types AS spend_type ON spend_type.id = spend.type_id")

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
		query = query.Where("make_date(month.year::int, month.month::int, day.day::int) BETWEEN ? AND ?",
			args.After, args.Before)
	case !args.After.IsZero():
		query = query.Where("make_date(month.year::int, month.month::int, day.day::int) >= ?", args.After)
	case !args.Before.IsZero():
		query = query.Where("make_date(month.year::int, month.month::int, day.day::int) <= ?", args.Before)
	}

	switch {
	case args.MinCost != 0 && args.MaxCost != 0:
		query = query.Where("spend.cost BETWEEN ? AND ?", args.MinCost, args.MaxCost)
	case args.MinCost != 0:
		query = query.Where("spend.cost >= ?", args.MinCost)
	case args.MaxCost != 0:
		query = query.Where("spend.cost <= ?", args.MaxCost)
	}

	query = query.WhereGroup(func(query *orm.Query) (*orm.Query, error) {
		typeIDs := args.TypeIDs
		for i, id := range typeIDs {
			if id == 0 {
				// Search for spends without type
				query = query.Where("spend.type_id IS NULL")
				typeIDs = append(typeIDs[:i], typeIDs[i+1:]...)
				break
			}
		}

		if len(typeIDs) != 0 {
			query = query.WhereOr("spend.type_id IN (?)", pg.In(typeIDs))
		}
		return query, nil
	})

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
	query = query.Order(orders...)

	return query
}
