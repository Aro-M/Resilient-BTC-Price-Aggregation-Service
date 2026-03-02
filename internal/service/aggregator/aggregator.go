package aggregator

import "sort"

type Result struct {
	Price       float64
	SourcesUsed int
	OK          bool
}

func Aggregate(prices []float64) Result {
	valid := make([]float64, 0, len(prices))
	for _, p := range prices {
		if p > 0 {
			valid = append(valid, p)
		}
	}
	if len(valid) == 0 {
		return Result{}
	}
	sort.Float64s(valid)
	n := len(valid)
	var median float64
	if n%2 == 0 {
		median = (valid[n/2-1] + valid[n/2]) / 2
	} else {
		median = valid[n/2]
	}
	return Result{Price: median, SourcesUsed: n, OK: true}
}