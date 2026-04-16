package calc

import (
	"testing"
	"time"
)

// --- computeTrend ---

func TestComputeTrendInsufficientData(t *testing.T) {
	windows := []WindowSummary{
		{Summary: Summary{Count: 1, Average: 100}},
	}
	if got := computeTrend(windows); got != TrendInsufficientData {
		t.Fatalf("expected insufficient_data, got %s", got)
	}
}

func TestComputeTrendUp(t *testing.T) {
	windows := []WindowSummary{
		{Summary: Summary{Count: 2, Average: 100}},
		{Summary: Summary{Count: 2, Average: 150}},
	}
	if got := computeTrend(windows); got != TrendUp {
		t.Fatalf("expected up, got %s", got)
	}
}

func TestComputeTrendDown(t *testing.T) {
	windows := []WindowSummary{
		{Summary: Summary{Count: 2, Average: 200}},
		{Summary: Summary{Count: 2, Average: 100}},
	}
	if got := computeTrend(windows); got != TrendDown {
		t.Fatalf("expected down, got %s", got)
	}
}

func TestComputeTrendFlat(t *testing.T) {
	windows := []WindowSummary{
		{Summary: Summary{Count: 2, Average: 100}},
		{Summary: Summary{Count: 2, Average: 102}},
	}
	if got := computeTrend(windows); got != TrendFlat {
		t.Fatalf("expected flat, got %s", got)
	}
}

// --- detectAlerts ---

func TestDetectAlertsNoAlerts(t *testing.T) {
	windows := []WindowSummary{
		{Summary: Summary{Count: 2, Average: 100, StdDev: 5}},
		{Summary: Summary{Count: 2, Average: 110, StdDev: 6}},
	}
	alerts := detectAlerts(windows)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d: %+v", len(alerts), alerts)
	}
}

func TestDetectAlertsSpike(t *testing.T) {
	// Window averages: 10, 10, 500 -> overall avg = ~176.67
	// 500 > 176.67*2 = 353.34 -> spike detected
	windows := []WindowSummary{
		{WindowStart: "2026-04-01T10:00:00Z", Summary: Summary{Count: 2, Average: 10, StdDev: 1}},
		{WindowStart: "2026-04-01T10:05:00Z", Summary: Summary{Count: 2, Average: 10, StdDev: 1}},
		{WindowStart: "2026-04-01T10:10:00Z", Summary: Summary{Count: 2, Average: 500, StdDev: 1}},
	}
	alerts := detectAlerts(windows)
	found := false
	for _, a := range alerts {
		if a.Type == AlertSpike {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spike alert, got %+v", alerts)
	}
}

// --- BuildReport ---

func TestBuildReportBasic(t *testing.T) {
	events := []MultiEvent{
		{Timestamp: mustParseTime("2026-04-01T10:00:00Z"), Metric: "revenue", Value: 100, Dimension: "cn"},
		{Timestamp: mustParseTime("2026-04-01T10:01:00Z"), Metric: "revenue", Value: 200, Dimension: "cn"},
		{Timestamp: mustParseTime("2026-04-01T10:06:00Z"), Metric: "revenue", Value: 300, Dimension: "cn"},
	}

	report := BuildReport(events, "revenue", "", 5*time.Minute, false)

	if report.Metric != "revenue" {
		t.Fatalf("expected metric revenue, got %s", report.Metric)
	}
	if report.WindowSize != "5m" {
		t.Fatalf("expected window_size 5m, got %s", report.WindowSize)
	}
	if len(report.CurrentWindows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(report.CurrentWindows))
	}
	if report.OverallSummary.Count != 3 {
		t.Fatalf("expected overall count 3, got %d", report.OverallSummary.Count)
	}
	if report.Trend != TrendUp {
		t.Fatalf("expected trend up, got %s", report.Trend)
	}
}

func TestBuildReportWithDimensionFilter(t *testing.T) {
	events := []MultiEvent{
		{Timestamp: mustParseTime("2026-04-01T10:00:00Z"), Metric: "revenue", Value: 100, Dimension: "cn"},
		{Timestamp: mustParseTime("2026-04-01T10:01:00Z"), Metric: "revenue", Value: 200, Dimension: "us"},
		{Timestamp: mustParseTime("2026-04-01T10:02:00Z"), Metric: "revenue", Value: 300, Dimension: "cn"},
	}

	report := BuildReport(events, "revenue", "cn", 5*time.Minute, false)

	if report.OverallSummary.Count != 2 {
		t.Fatalf("expected 2 events for cn dimension, got %d", report.OverallSummary.Count)
	}
}

func TestBuildReportEmpty(t *testing.T) {
	report := BuildReport(nil, "revenue", "", 5*time.Minute, false)

	if report.OverallSummary.Count != 0 {
		t.Fatalf("expected 0 count, got %d", report.OverallSummary.Count)
	}
	if report.Trend != TrendInsufficientData {
		t.Fatalf("expected insufficient_data trend, got %s", report.Trend)
	}
	if len(report.Alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(report.Alerts))
	}
}
