package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/sirupsen/logrus"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

func (db DB) SearchSpends(ctx context.Context, args db_common.SearchSpendsArgs) ([]*db_common.Spend, error) {
	log := request_id.FromContextToLogger(ctx, db.log)
	log = log.WithFields(logrus.Fields{
		"title": args.Title, "notes": args.Notes, "after": args.After, "before": args.Before,
		"min_cost": args.MinCost, "max_cost": args.MaxCost, "type_ids": args.TypeIDs,
	})

	var pgSpends []*searchSpendsModel

	startTime := time.Now()
	err := db.db.RunInTransaction(func(tx *pg.Tx) error {
		query := db.buildSearchSpendsQuery(tx, args)
		return query.Select(&pgSpends)
	})
	if err != nil {
		log.WithError(err).Error("couldn't select Spends")
		return nil, err
	}

	// Convert the internal model to the common one
	spends := make([]*db_common.Spend, 0, len(pgSpends))
	for i := range pgSpends {
		s := &db_common.Spend{
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
			s.Type = &db_common.SpendType{
				ID:   pgSpends[i].Type.ID,
				Name: pgSpends[i].Type.Name,
			}
		}
		spends = append(spends, s)
	}

	log = log.WithFields(logrus.Fields{
		"time":         time.Since(startTime),
		"spend_number": len(spends),
	})

	log.Debug("return found Spends")
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
	Type  struct {
		ID   uint
		Name string
	}
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

	switch {
	case args.WithoutType:
		query = query.Where("spend.type_id IS NULL")
	case len(args.TypeIDs) != 0:
		query = query.WhereIn("spend.type_id IN (?)", args.TypeIDs)
	}

	return query
}
