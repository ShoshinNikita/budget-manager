package statistics

import (
	"sort"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type SpentByDayDataset []SpentByDayData

type SpentByDayData struct {
	Year  int        `json:"year"`
	Month time.Month `json:"month"`
	Day   int        `json:"day"`

	Spent money.Money `json:"spent"`
}

func CalculateSpentByDay(spends []db.Spend, startDate, endDate time.Time) SpentByDayDataset {
	spentByDay := make(map[time.Time]money.Money)
	for _, spend := range spends {
		t := time.Date(spend.Year, spend.Month, spend.Day, 0, 0, 0, 0, time.UTC)

		spent := spentByDay[t]
		spent = spent.Add(spend.Cost)
		spentByDay[t] = spent
	}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	for t := startDate; t.Before(endDate) || t.Equal(endDate); t = t.Add(24 * time.Hour) {
		if _, ok := spentByDay[t]; !ok {
			spentByDay[t] = 0
		}
	}

	res := make(SpentByDayDataset, 0, len(spentByDay))
	for t, spent := range spentByDay {
		res = append(res, SpentByDayData{
			Year:  t.Year(),
			Month: t.Month(),
			Day:   t.Day(),
			//
			Spent: spent,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Year != res[j].Year {
			return res[i].Year < res[j].Year
		}
		if res[i].Month != res[j].Month {
			return res[i].Month < res[j].Month
		}
		return res[i].Day < res[j].Day
	})

	return res
}
