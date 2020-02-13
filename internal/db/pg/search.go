package pg

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// buildSearchSpendsQuery builds query to search spends
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
//   WHERE make_date(month.year::int, month.month::int, day.day::int) BETWEEN ':after'::date AND ':before'::date
//         AND title LIKE ':pattern'
//         AND spend.cost BETWEEN :min AND :max
//         AND spend.type_id IN (:type_ids)
//
//   ORDER BY month.year, month.month, day.day, spend.id;
//
func (DB) buildSearchSpendsQuery(tx *pg.Tx, args db_common.SearchSpendsArgs) *orm.Query {
	query := tx.Model((*Spend)(nil)).
		ColumnExpr("spend.id AS id").
		ColumnExpr("month.year AS year").
		ColumnExpr("month.month AS month").
		ColumnExpr("day.day AS day").
		ColumnExpr("spend.title AS title").
		ColumnExpr("spend.notes AS notes").
		ColumnExpr("spend.cost AS cost").
		ColumnExpr("spend_type.id AS type__id").
		ColumnExpr("spend_type.name AS type__name").
		//
		Join("INNER JOIN days AS day ON day.id = spend.day_id").
		Join("INNER JOIN months AS month ON month.id = day.month_id").
		Join("LEFT JOIN spend_types AS spend_type ON spend_type.id = spend.type_id").
		//
		Order("month.year").Order("month.month").Order("day.day").Order("spend.id")

	if args.Title != "" {
		query = query.Where("spend.title LIKE ?", args.Title)
	}

	if args.Notes != "" {
		query = query.Where("spend.notes LIKE ?", args.Notes)
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

	if len(args.TypeIDs) != 0 {
		query = query.WhereIn("spend.type_id IN (?)", args.TypeIDs)
	}

	return query
}
