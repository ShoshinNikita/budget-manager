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

func CalculateCostIntervals(spends []db.Spend, maxIntervalNumber int) (intervals []CostInterval) {
	if len(spends) == 0 {
		return nil
	}

	costs := extractSortedCosts(spends)
	for intervalNumber := maxIntervalNumber; intervalNumber > 0; intervalNumber-- {
		intervals = prepareIntervals(costs, intervalNumber)
		intervals = fillIntervals(costs, intervals)

		if last := intervals[len(intervals)-1]; last.From >= last.To {
			// Sometimes there can be a situation when 'from' >= 'to' because we round the interval. Just try
			// we the lower interval
			continue
		}

		var c int
		for i := range intervals {
			if intervals[i].Total != 0 {
				c++
			}
		}
		if c > len(intervals)/2 {
			break
		}
	}

	// Round interval totals for more beautiful view
	for i := range intervals {
		intervals[i].Total = intervals[i].Total.Round()
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

// prepareIntervals prepares cost intervals excluding costs less than p5 and greater than p95 (p - percentile)
func prepareIntervals(costs []money.Money, intervalNumber int) []CostInterval {
	min := getPercentileValue(costs, 5).Floor()
	max := getPercentileValue(costs, 95).Ceil()

	delta := max.Sub(min)
	interval := delta.Div(int64(intervalNumber)).Round()

	intervals := make([]CostInterval, 0, intervalNumber)
	next := min
	for i := 0; i < intervalNumber; i++ {
		from := next
		next = next.Add(interval)
		to := biggestValueBefore(next)
		if i+1 == intervalNumber {
			to = max
		}

		intervals = append(intervals, CostInterval{From: from, To: to})
	}

	return intervals
}

// getPercentileValue returns a value at the nth percentile. It uses the nearest rank method to
// find the percentile rank - https://en.wikipedia.org/wiki/Percentile#The_nearest-rank_method
func getPercentileValue(costs []money.Money, n int) money.Money {
	i := float64(n) / 100 * float64(len(costs))
	index := int(math.Ceil(i)) - 1
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

func fillIntervals(costs []money.Money, intervals []CostInterval) []CostInterval {
	for _, cost := range costs {
		for i := range intervals {
			if intervals[i].From <= cost && cost <= intervals[i].To {
				intervals[i].Count++
				intervals[i].Total = intervals[i].Total.Add(cost)
				break
			}
		}
	}
	return intervals
}
