package calc

import (
	"errors"
	"math"
	"sort"
)

// Summary holds descriptive statistics for a slice of float64 values.
type Summary struct {
	Count   int     `json:"count"`
	Sum     float64 `json:"sum"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
	Median  float64 `json:"median"`
	Range   float64 `json:"range"`
}

// Sum returns the sum of all values. Returns 0 for an empty slice.
func Sum(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}

// Average returns the arithmetic mean of values. Returns 0 for an empty slice.
func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return Sum(values) / float64(len(values))
}

// Median returns the median of values without mutating the input slice.
// Returns 0 for an empty slice.
func Median(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}

// PercentChange calculates the percentage change from prev to current.
// Returns an error if prev is zero.
func PercentChange(prev, current float64) (float64, error) {
	if prev == 0 {
		return 0, errors.New("previous value cannot be zero")
	}
	return ((current - prev) / math.Abs(prev)) * 100, nil
}

// BuildSummary computes descriptive statistics for the given values.
// Returns a zero-value Summary for an empty slice.
func BuildSummary(values []float64) Summary {
	n := len(values)
	if n == 0 {
		return Summary{}
	}

	s := Summary{
		Count:   n,
		Sum:     Sum(values),
		Average: Average(values),
		Median:  Median(values),
		Min:     values[0],
		Max:     values[0],
	}

	for _, v := range values[1:] {
		if v < s.Min {
			s.Min = v
		}
		if v > s.Max {
			s.Max = v
		}
	}

	s.Range = s.Max - s.Min
	return s
}
