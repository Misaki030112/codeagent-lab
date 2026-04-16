package calc

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestFixtureBasicCSV verifies that the shared basic.csv fixture produces expected results.
func TestFixtureBasicCSV(t *testing.T) {
	f, err := os.Open("../fixtures/events/basic.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	events, warnings := ParseMultiCSV(f)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(events) != 6 {
		t.Fatalf("expected 6 events, got %d", len(events))
	}

	// All events are revenue
	for _, e := range events {
		if e.Metric != "revenue" {
			t.Fatalf("expected metric revenue, got %s", e.Metric)
		}
	}

	report := BuildReport(events, "revenue", "", 5*time.Minute, false)
	if len(report.CurrentWindows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(report.CurrentWindows))
	}
	if report.OverallSummary.Count != 6 {
		t.Fatalf("expected overall count 6, got %d", report.OverallSummary.Count)
	}

	// Verify window 1
	w0 := report.CurrentWindows[0]
	if w0.WindowStart != "2026-04-01T10:00:00Z" {
		t.Fatalf("window 0 start = %s, want 2026-04-01T10:00:00Z", w0.WindowStart)
	}
	if w0.Summary.Count != 4 {
		t.Fatalf("window 0 count = %d, want 4", w0.Summary.Count)
	}

	// Verify window 2
	w1 := report.CurrentWindows[1]
	if w1.WindowStart != "2026-04-01T10:05:00Z" {
		t.Fatalf("window 1 start = %s, want 2026-04-01T10:05:00Z", w1.WindowStart)
	}
	if w1.Summary.Count != 2 {
		t.Fatalf("window 1 count = %d, want 2", w1.Summary.Count)
	}

	// Overall summary checks
	if !almostEqual(report.OverallSummary.Sum, 756.0) {
		t.Fatalf("overall sum = %f, want 756.0", report.OverallSummary.Sum)
	}
	if !almostEqual(report.OverallSummary.Min, 80.0) {
		t.Fatalf("overall min = %f, want 80.0", report.OverallSummary.Min)
	}
	if !almostEqual(report.OverallSummary.Max, 200.0) {
		t.Fatalf("overall max = %f, want 200.0", report.OverallSummary.Max)
	}
}

// TestFixtureMultiMetricCSV verifies multi-metric fixture with filtering.
func TestFixtureMultiMetricCSV(t *testing.T) {
	f, err := os.Open("../fixtures/events/multi_metric.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	events, warnings := ParseMultiCSV(f)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(events) != 9 {
		t.Fatalf("expected 9 events, got %d", len(events))
	}

	// Filter by revenue
	revenueReport := BuildReport(events, "revenue", "", 5*time.Minute, false)
	if revenueReport.OverallSummary.Count != 5 {
		t.Fatalf("revenue count = %d, want 5", revenueReport.OverallSummary.Count)
	}

	// Filter by latency_ms
	latencyReport := BuildReport(events, "latency_ms", "", 5*time.Minute, false)
	if latencyReport.OverallSummary.Count != 4 {
		t.Fatalf("latency count = %d, want 4", latencyReport.OverallSummary.Count)
	}
}

// TestFixtureInvalidRowsCSV verifies that invalid rows produce structured warnings.
func TestFixtureInvalidRowsCSV(t *testing.T) {
	f, err := os.Open("../fixtures/events/invalid_rows.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	events, warnings := ParseMultiCSV(f)

	// Should have valid events (rows 1, 5, 7)
	if len(events) != 3 {
		t.Fatalf("expected 3 valid events, got %d", len(events))
	}

	// Should have warnings for invalid rows
	if len(warnings) < 3 {
		t.Fatalf("expected at least 3 warnings, got %d", len(warnings))
	}

	// Verify warnings have proper row numbers
	for _, w := range warnings {
		if w.Row < 2 {
			t.Fatalf("warning row should be >= 2 (after header), got %d", w.Row)
		}
	}
}

// TestFixtureEmptyCSV verifies behavior with header-only CSV.
func TestFixtureEmptyCSV(t *testing.T) {
	f, err := os.Open("../fixtures/events/empty.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	events, warnings := ParseMultiCSV(f)
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	report := BuildReport(events, "revenue", "", 5*time.Minute, false)
	if report.OverallSummary.Count != 0 {
		t.Fatalf("expected 0 count for empty input, got %d", report.OverallSummary.Count)
	}
}

// TestFixtureSingleValue verifies behavior with a single event.
func TestFixtureSingleValue(t *testing.T) {
	f, err := os.Open("../fixtures/events/single_metric_single_value.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	events, warnings := ParseMultiCSV(f)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	report := BuildReport(events, "revenue", "", 5*time.Minute, false)
	if report.OverallSummary.Count != 1 {
		t.Fatalf("expected count 1, got %d", report.OverallSummary.Count)
	}
	if !almostEqual(report.OverallSummary.Sum, 100.0) {
		t.Fatalf("expected sum 100.0, got %f", report.OverallSummary.Sum)
	}
	if report.Trend != TrendInsufficientData {
		t.Fatalf("expected insufficient_data trend for single window, got %s", report.Trend)
	}
}

// TestCrossLanguageContractBasicSum verifies the basic sum value matches the Python expected output.
func TestCrossLanguageContractBasicSum(t *testing.T) {
	// These are the contract values that Python must also produce
	values := []float64{120.5, 80.0, 95.5, 110.0, 200.0, 150.0}

	s := BuildSummary(values)

	if s.Count != 6 {
		t.Fatalf("count = %d, want 6", s.Count)
	}
	if !almostEqual(s.Sum, 756.0) {
		t.Fatalf("sum = %f, want 756.0", s.Sum)
	}
	if !almostEqual(s.Min, 80.0) {
		t.Fatalf("min = %f, want 80.0", s.Min)
	}
	if !almostEqual(s.Max, 200.0) {
		t.Fatalf("max = %f, want 200.0", s.Max)
	}
	if !almostEqual(s.Average, 126.0) {
		t.Fatalf("average = %f, want 126.0", s.Average)
	}

	// Verify variance and std_dev are reasonable
	if s.Variance <= 0 {
		t.Fatalf("variance should be > 0, got %f", s.Variance)
	}
	if s.StdDev <= 0 {
		t.Fatalf("std_dev should be > 0, got %f", s.StdDev)
	}

	// P90 and P95 should be between max and average
	if s.P90 < s.Average || s.P90 > s.Max {
		t.Fatalf("P90 = %f should be between average (%f) and max (%f)", s.P90, s.Average, s.Max)
	}
}

// TestParseMultiCSVFromString is a quick sanity check for parsing from string.
func TestParseMultiCSVFromString(t *testing.T) {
	input := "timestamp,metric,value,dimension,source\n2026-04-01T10:00:00Z,revenue,100.0,cn,ads\n"
	events, warnings := ParseMultiCSV(strings.NewReader(input))
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Metric != "revenue" || events[0].Value != 100.0 {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}
