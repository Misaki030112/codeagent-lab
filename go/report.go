package calc

import (
	"sort"
	"time"
)

// Thresholds for trend detection and anomaly alerting.
const (
	// trendThresholdPct is the minimum percent change to classify as up/down trend.
	trendThresholdPct = 5.0
	// spikeMultiplier defines the ratio above overall average that triggers a spike alert.
	spikeMultiplier = 2.0
	// dropMultiplier defines the ratio below overall average that triggers a drop alert.
	dropMultiplier = 0.5
	// varianceMultiplier defines the ratio above overall std_dev that triggers a high-variance alert.
	varianceMultiplier = 2.0
)

// BuildReport generates a full analytics report for filtered events.
func BuildReport(events []MultiEvent, metric, dimension string, ws time.Duration, fillEmpty bool) Report {
	filtered := FilterEvents(events, metric, dimension)

	// Sort by timestamp
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})

	windows := BuildMultiWindowSummaries(filtered, ws, fillEmpty)
	if windows == nil {
		windows = []WindowSummary{}
	}

	// Collect all values in time order for overall summary
	values := make([]float64, len(filtered))
	for i, e := range filtered {
		values[i] = e.Value
	}
	overallSummary := BuildSummaryOrdered(values, values)

	trend := computeTrend(windows)
	alerts := detectAlerts(windows)

	return Report{
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Metric:          metric,
		Dimension:       dimension,
		WindowSize:      FormatWindowSize(ws),
		CurrentWindows:  windows,
		PreviousWindows: []WindowSummary{},
		OverallSummary:  overallSummary,
		Trend:           trend,
		Alerts:          alerts,
	}
}

// computeTrend determines the overall trend from window summaries.
func computeTrend(windows []WindowSummary) Trend {
	if len(windows) < 2 {
		return TrendInsufficientData
	}

	first := windows[0].Summary.Average
	last := windows[len(windows)-1].Summary.Average

	if first == 0 && last == 0 {
		return TrendFlat
	}

	var pctChange float64
	if first != 0 {
		pctChange = ((last - first) / first) * 100
	} else {
		// first is 0, last is not: treat as up or down based on sign
		if last > 0 {
			return TrendUp
		}
		return TrendDown
	}

	if pctChange > trendThresholdPct {
		return TrendUp
	}
	if pctChange < -trendThresholdPct {
		return TrendDown
	}
	return TrendFlat
}

// detectAlerts checks each window for anomalies using heuristic rules.
// Spike: window average > 2x overall average
// Drop: window average < 0.5x overall average
// High variance: window std_dev > 2x overall std_dev
func detectAlerts(windows []WindowSummary) []Alert {
	if len(windows) < 2 {
		return []Alert{}
	}

	// Compute overall average and std_dev across all windows
	var allValues []float64
	for _, w := range windows {
		if w.Summary.Count > 0 {
			allValues = append(allValues, w.Summary.Average)
		}
	}
	if len(allValues) == 0 {
		return []Alert{}
	}

	overallAvg := Average(allValues)
	overallStdDev := StdDev(allValues)

	// Thresholds are constant across all windows — compute once.
	spikeThreshold := overallAvg * spikeMultiplier
	dropThreshold := overallAvg * dropMultiplier
	varThreshold := overallStdDev * varianceMultiplier

	var alerts []Alert
	for _, w := range windows {
		if w.Summary.Count == 0 {
			continue
		}

		if overallAvg > 0 && w.Summary.Average > spikeThreshold {
			alerts = append(alerts, Alert{
				Type:        AlertSpike,
				WindowStart: w.WindowStart,
				Message:     "window average is more than 2x the overall average",
				Value:       w.Summary.Average,
				Threshold:   spikeThreshold,
			})
		}

		if overallAvg > 0 && w.Summary.Average < dropThreshold {
			alerts = append(alerts, Alert{
				Type:        AlertDrop,
				WindowStart: w.WindowStart,
				Message:     "window average is less than 0.5x the overall average",
				Value:       w.Summary.Average,
				Threshold:   dropThreshold,
			})
		}

		if overallStdDev > 0 && w.Summary.StdDev > varThreshold {
			alerts = append(alerts, Alert{
				Type:        AlertHighVariance,
				WindowStart: w.WindowStart,
				Message:     "window standard deviation is more than 2x the overall",
				Value:       w.Summary.StdDev,
				Threshold:   varThreshold,
			})
		}
	}

	return alerts
}
