package base

import (
	"time"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type Day struct {
	ID      uint        `db:"id"`
	MonthID uint        `db:"month_id"`
	Day     int         `db:"day"`
	Saldo   money.Money `db:"saldo"` // DailyBudget - Cost of all Spends

	Spends []Spend `db:"-"`
}

// ToCommon converts Day to common Day structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (d Day) ToCommon(year int, month time.Month) common.Day {
	return common.Day{
		ID:    d.ID,
		Year:  year,
		Month: month,
		Day:   d.Day,
		Saldo: d.Saldo,
		Spends: func() []common.Spend {
			spends := make([]common.Spend, 0, len(d.Spends))
			for i := range d.Spends {
				spends = append(spends, d.Spends[i].ToCommon(year, month, d.Day))
			}
			return spends
		}(),
	}
}
