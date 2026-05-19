package calc

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// --- Sum ---

func TestSumEmpty(t *testing.T) {
	if got := Sum(nil); got != 0 {
		t.Fatalf("Sum(nil) = %f, want 0", got)
	}
}

func TestSumSingle(t *testing.T) {
	if got := Sum([]float64{42}); got != 42 {
		t.Fatalf("Sum([42]) = %f, want 42", got)
	}
}

func TestSumMultiple(t *testing.T) {
	got := Sum([]float64{1, 2, 3, 4, 5})
	if got != 15 {
		t.Fatalf("Sum([1..5]) = %f, want 15", got)
	}
}

func TestSumNegatives(t *testing.T) {
	got := Sum([]float64{-3, -7, 10})
	if got != 0 {
		t.Fatalf("Sum([-3,-7,10]) = %f, want 0", got)
	}
}

func TestSumDecimals(t *testing.T) {
	got := Sum([]float64{0.1, 0.2, 0.3})
	if !almostEqual(got, 0.6) {
		t.Fatalf("Sum([0.1,0.2,0.3]) = %f, want ~0.6", got)
	}
}

// --- Average ---

func TestAverageEmpty(t *testing.T) {
	if got := Average(nil); got != 0 {
		t.Fatalf("Average(nil) = %f, want 0", got)
	}
}

func TestAverageSingle(t *testing.T) {
	if got := Average([]float64{7}); got != 7 {
		t.Fatalf("Average([7]) = %f, want 7", got)
	}
}

func TestAverageMultiple(t *testing.T) {
	got := Average([]float64{2, 4, 6})
	if got != 4 {
		t.Fatalf("Average([2,4,6]) = %f, want 4", got)
	}
}

func TestAverageNegatives(t *testing.T) {
	got := Average([]float64{-10, 10})
	if got != 0 {
		t.Fatalf("Average([-10,10]) = %f, want 0", got)
	}
}

// --- Median ---

func TestMedianEmpty(t *testing.T) {
	if got := Median(nil); got != 0 {
		t.Fatalf("Median(nil) = %f, want 0", got)
	}
}

func TestMedianSingle(t *testing.T) {
	if got := Median([]float64{5}); got != 5 {
		t.Fatalf("Median([5]) = %f, want 5", got)
	}
}

func TestMedianOdd(t *testing.T) {
	got := Median([]float64{3, 1, 2})
	if got != 2 {
		t.Fatalf("Median([3,1,2]) = %f, want 2", got)
	}
}

func TestMedianEven(t *testing.T) {
	got := Median([]float64{4, 1, 3, 2})
	if got != 2.5 {
		t.Fatalf("Median([4,1,3,2]) = %f, want 2.5", got)
	}
}

func TestMedianDuplicates(t *testing.T) {
	got := Median([]float64{5, 5, 5, 5})
	if got != 5 {
		t.Fatalf("Median([5,5,5,5]) = %f, want 5", got)
	}
}

func TestMedianDoesNotMutateInput(t *testing.T) {
	input := []float64{3, 1, 2}
	Median(input)
	if input[0] != 3 || input[1] != 1 || input[2] != 2 {
		t.Fatalf("Median mutated input slice: %v", input)
	}
}

func TestMedianNegatives(t *testing.T) {
	got := Median([]float64{-5, -1, -3})
	if got != -3 {
		t.Fatalf("Median([-5,-1,-3]) = %f, want -3", got)
	}
}

// --- PercentChange ---

func TestPercentChangePositive(t *testing.T) {
	got, err := PercentChange(100, 150)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 50 {
		t.Fatalf("PercentChange(100,150) = %f, want 50", got)
	}
}

func TestPercentChangeNegativeDecrease(t *testing.T) {
	got, err := PercentChange(200, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -50 {
		t.Fatalf("PercentChange(200,100) = %f, want -50", got)
	}
}

func TestPercentChangeZeroPrev(t *testing.T) {
	_, err := PercentChange(0, 50)
	if err == nil {
		t.Fatal("PercentChange(0, 50) expected error, got nil")
	}
}

func TestPercentChangeNegativePrev(t *testing.T) {
	got, err := PercentChange(-100, -50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 50 {
		t.Fatalf("PercentChange(-100,-50) = %f, want 50", got)
	}
}

func TestPercentChangeNoChange(t *testing.T) {
	got, err := PercentChange(42, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("PercentChange(42,42) = %f, want 0", got)
	}
}

// --- BuildSummary ---

func TestBuildSummaryEmpty(t *testing.T) {
	s := BuildSummary(nil)
	if s.Count != 0 || s.Sum != 0 || s.Min != 0 || s.Max != 0 || s.Average != 0 || s.Median != 0 {
		t.Fatalf("BuildSummary(nil) = %+v, want zero", s)
	}
}

func TestBuildSummarySingle(t *testing.T) {
	s := BuildSummary([]float64{10})
	if s.Count != 1 || s.Sum != 10 || s.Min != 10 || s.Max != 10 || s.Average != 10 || s.Median != 10 {
		t.Fatalf("BuildSummary([10]) = %+v", s)
	}
}

func TestBuildSummaryMultiple(t *testing.T) {
	s := BuildSummary([]float64{1, 2, 3, 4, 5})
	if s.Count != 5 {
		t.Fatalf("Count = %d, want 5", s.Count)
	}
	if s.Sum != 15 {
		t.Fatalf("Sum = %f, want 15", s.Sum)
	}
	if s.Min != 1 {
		t.Fatalf("Min = %f, want 1", s.Min)
	}
	if s.Max != 5 {
		t.Fatalf("Max = %f, want 5", s.Max)
	}
	if s.Average != 3 {
		t.Fatalf("Average = %f, want 3", s.Average)
	}
	if s.Median != 3 {
		t.Fatalf("Median = %f, want 3", s.Median)
	}
}

func TestBuildSummaryNegatives(t *testing.T) {
	s := BuildSummary([]float64{-10, -5, 0, 5, 10})
	if s.Min != -10 {
		t.Fatalf("Min = %f, want -10", s.Min)
	}
	if s.Max != 10 {
		t.Fatalf("Max = %f, want 10", s.Max)
	}
	if s.Average != 0 {
		t.Fatalf("Average = %f, want 0", s.Average)
	}
}

func TestBuildSummaryDuplicates(t *testing.T) {
	s := BuildSummary([]float64{7, 7, 7})
	if s.Min != 7 || s.Max != 7 || s.Average != 7 || s.Median != 7 {
		t.Fatalf("BuildSummary([7,7,7]) = %+v", s)
	}
	if s.Variance != 0 {
		t.Fatalf("Variance([7,7,7]) = %f, want 0", s.Variance)
	}
	if s.StdDev != 0 {
		t.Fatalf("StdDev([7,7,7]) = %f, want 0", s.StdDev)
	}
}

// --- Variance ---

func TestVariance(t *testing.T) {
	tests := []struct {
		name   string
		input  []float64
		wantV  float64
		wantSD float64
	}{
		{"empty", nil, 0, 0},
		{"single", []float64{5}, 0, 0},
		{"two equal", []float64{3, 3}, 0, 0},
		{"two different", []float64{2, 4}, 1, 1},
		{"positive integers", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 4, 2},
		{"negatives", []float64{-3, -1, 1, 3}, 5, math.Sqrt(5)},
		{"all same", []float64{10, 10, 10, 10}, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Variance(tc.input); !almostEqual(got, tc.wantV) {
				t.Errorf("Variance(%v) = %f, want %f", tc.input, got, tc.wantV)
			}
			if got := StdDev(tc.input); !almostEqual(got, tc.wantSD) {
				t.Errorf("StdDev(%v) = %f, want %f", tc.input, got, tc.wantSD)
			}
		})
	}
}

func TestBuildSummaryIncludesVarianceAndStdDev(t *testing.T) {
	s := BuildSummary([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if !almostEqual(s.Variance, 4) {
		t.Fatalf("Variance = %f, want 4", s.Variance)
	}
	if !almostEqual(s.StdDev, 2) {
		t.Fatalf("StdDev = %f, want 2", s.StdDev)
	}
}

func TestBuildSummaryEmptyVariance(t *testing.T) {
	s := BuildSummary(nil)
	if s.Variance != 0 || s.StdDev != 0 {
		t.Fatalf("BuildSummary(nil) variance/stddev not zero: %+v", s)
	}
}
