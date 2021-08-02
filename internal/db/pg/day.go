package pg

import (
	"time"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// Day represents day entity in PostgreSQL db
type Day struct {
	tableName struct{} `pg:"days"`

	ID uint `pg:"id,pk"`

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	Day int `pg:"day"`
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  money.Money `pg:"saldo,use_zero"`
	Spends []Spend     `pg:"rel:has-many,join_fk:day_id"`
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
