package statistics

import (
	"math"
	"sort"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type CostInterval struct {
	From money.Money `json:"from"`
	To   money.Money `json:"to"`

	Count int         `json:"count"`
	Total money.Money `json:"total"`
}

func CalculateCostIntervals(spends []db.Spend, intervalNumber int) []CostInterval {
	if len(spends) == 0 {
		return nil
	}

	costs := extractSortedCosts(spends)
	intervals := prepareIntervals(costs, intervalNumber)

	// Fill intervals
	for _, s := range spends {
		for i := range intervals {
			if intervals[i].From <= s.Cost && s.Cost <= intervals[i].To {
				intervals[i].Count++
				intervals[i].Total = intervals[i].Total.Add(s.Cost)
				break
			}
		}
	}

	return intervals
}

// extractSortedCosts returns a sorted slice of Spend costs
func extractSortedCosts(spends []db.Spend) []money.Money {
	res := make([]money.Money, 0, len(spends))
	for _, s := range spends {
		res = append(res, s.Cost)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

// prepareIntervals prepares cost intervals. It use delta between of 95th and 5th percentile values to calculate
// intervals
func prepareIntervals(costs []money.Money, intervalNumber int) []CostInterval {
	min := getPercentileValue(costs, 5)
	max := getPercentileValue(costs, 95)

	delta := max.Sub(min).Float()
	// Divide by intervalNumber-2 because there are 2 additional intervals: [min, p5Min), [p95Max, max]
	interval := int64(math.Ceil(delta / float64(intervalNumber-2)))

	intervals := make([]CostInterval, intervalNumber)
	// Set the first interval - [0, 5pMin)
	intervals[0] = CostInterval{
		From: costs[0], To: biggestValueBefore(min),
	}
	next := min
	for i := 1; i < intervalNumber-1; i++ {
		from := next
		next = next.Add(money.FromInt(interval))
		to := next

		intervals[i] = CostInterval{
			From: from,
			To:   biggestValueBefore(to),
		}
	}
	// Set the last interval - [95pMax, max]
	intervals[intervalNumber-1] = CostInterval{
		From: next, To: costs[len(costs)-1],
	}

	return intervals
}

// getPercentileValue returns a value at the nth percentile. It uses the nearest rank method to
// find the percentile rank - https://en.wikipedia.org/wiki/Percentile#The_nearest-rank_method
func getPercentileValue(costs []money.Money, n float64) money.Money {
	i := n / 100 * float64(len(costs))
	index := int(math.Ceil(i))
	switch {
	case index < 0:
		index = 0
	case index >= len(costs):
		index = len(costs) - 1
	}

	return costs[index]
}

// biggestValueBefore returns the biggest value before 'm'. It can be used to represent open intervals - (a, b)
// For example, if max money precision is 2, it will return 'm-0.01'. If precision is 3 - 'm-0.001' and so one
func biggestValueBefore(m money.Money) money.Money {
	return m - 1
}