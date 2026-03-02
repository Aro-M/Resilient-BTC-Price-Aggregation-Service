package aggregator

import (
	"math"
	"testing"
)

func TestAggregate(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		expected float64
		ok       bool
	}{
		{
			name:     "single valid price",
			prices:   []float64{50000.0},
			expected: 50000.0,
			ok:       true,
		},
		{
			name:     "odd number of prices (median is middle)",
			prices:   []float64{50000.0, 52000.0, 48000.0},
			expected: 50000.0,
			ok:       true,
		},
		{
			name:     "even number of prices (median is average of middle two)",
			prices:   []float64{50000.0, 52000.0, 48000.0, 54000.0},
			expected: 51000.0,
			ok:       true,
		},
		{
			name:     "prices with invalid (zero or negative) values",
			prices:   []float64{50000.0, 0, -1000.0, 52000.0},
			expected: 51000.0,
			ok:       true,
		},
		{
			name:     "all invalid prices",
			prices:   []float64{0, -100.0},
			expected: 0,
			ok:       false,
		},
		{
			name:     "empty prices array",
			prices:   []float64{},
			expected: 0,
			ok:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := Aggregate(tc.prices)
			if res.OK != tc.ok {
				t.Fatalf("expected OK=%v, got OK=%v", tc.ok, res.OK)
			}
			if tc.ok && math.Abs(res.Price-tc.expected) > 0.001 {
				t.Fatalf("expected price %v, got %v", tc.expected, res.Price)
			}
		})
	}
}
