package calc

import (
	"errors"
	"math"
	"sort"
)

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

// Variance returns the population variance of values. Returns 0 for an empty slice.
func Variance(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	avg := Average(values)
	var sumSq float64
	for _, v := range values {
		d := v - avg
		sumSq += d * d
	}
	return sumSq / float64(n)
}

// StdDev returns the population standard deviation of values. Returns 0 for an empty slice.
func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}

// Percentile returns the p-th percentile (0–100) using linear interpolation.
// Returns 0 for an empty slice.
func Percentile(values []float64, p float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	if n == 1 {
		return sorted[0]
	}

	rank := (p / 100.0) * float64(n-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper {
		return sorted[lower]
	}
	frac := rank - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
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
// The first and last parameters represent the time-ordered first and last values.
// Returns a zero-value Summary for an empty slice.
func BuildSummary(values []float64) Summary {
	return BuildSummaryOrdered(values, values)
}

// BuildSummaryOrdered computes descriptive statistics for the given values.
// orderedValues provides the time-ordered sequence for first/last/delta/percent_change.
// If orderedValues is nil or empty, values is used for ordering as well.
func BuildSummaryOrdered(values []float64, orderedValues []float64) Summary {
	n := len(values)
	if n == 0 {
		return Summary{}
	}

	if len(orderedValues) == 0 {
		orderedValues = values
	}

	s := Summary{
		Count:    n,
		Sum:      Sum(values),
		Average:  Average(values),
		Median:   Median(values),
		Variance: Variance(values),
		StdDev:   StdDev(values),
		P90:      Percentile(values, 90),
		P95:      Percentile(values, 95),
		Min:      values[0],
		Max:      values[0],
		First:    orderedValues[0],
		Last:     orderedValues[len(orderedValues)-1],
	}

	for _, v := range values[1:] {
		if v < s.Min {
			s.Min = v
		}
		if v > s.Max {
			s.Max = v
		}
	}

	s.Delta = s.Last - s.First
	if s.First != 0 {
		pc, _ := PercentChange(s.First, s.Last)
		s.PercentChange = &pc
	}

	return s
}
