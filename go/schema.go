package calc

import "time"

// Summary holds descriptive statistics for a set of values.
// All fields use the shared analytics contract (see fixtures/schema/report_schema.json).
type Summary struct {
	Count         int      `json:"count"`
	Sum           float64  `json:"sum"`
	Min           float64  `json:"min"`
	Max           float64  `json:"max"`
	Average       float64  `json:"average"`
	Median        float64  `json:"median"`
	Variance      float64  `json:"variance"`
	StdDev        float64  `json:"std_dev"`
	P90           float64  `json:"p90"`
	P95           float64  `json:"p95"`
	First         float64  `json:"first"`
	Last          float64  `json:"last"`
	Delta         float64  `json:"delta"`
	PercentChange *float64 `json:"percent_change"`
}

// MultiEvent represents a single timestamped observation with metric, dimension, and source.
type MultiEvent struct {
	Timestamp time.Time
	Metric    string
	Value     float64
	Dimension string
	Source    string
}

// WindowSummary holds the time bounds and statistics for one window.
type WindowSummary struct {
	WindowStart string  `json:"window_start"`
	WindowEnd   string  `json:"window_end"`
	Summary     Summary `json:"summary"`
}

// Trend represents the overall direction of a metric.
type Trend string

const (
	TrendUp               Trend = "up"
	TrendDown             Trend = "down"
	TrendFlat             Trend = "flat"
	TrendInsufficientData Trend = "insufficient_data"
)

// AlertType categorizes anomaly detections.
type AlertType string

const (
	AlertSpike        AlertType = "spike"
	AlertDrop         AlertType = "drop"
	AlertHighVariance AlertType = "high_variance"
)

// Alert represents a detected anomaly in a window.
type Alert struct {
	Type        AlertType `json:"type"`
	WindowStart string    `json:"window_start"`
	Message     string    `json:"message"`
	Value       float64   `json:"value,omitempty"`
	Threshold   float64   `json:"threshold,omitempty"`
}

// Report is the top-level analytics report output.
type Report struct {
	GeneratedAt     string          `json:"generated_at"`
	Metric          string          `json:"metric"`
	Dimension       string          `json:"dimension"`
	WindowSize      string          `json:"window_size"`
	CurrentWindows  []WindowSummary `json:"current_windows"`
	PreviousWindows []WindowSummary `json:"previous_windows"`
	OverallSummary  Summary         `json:"overall_summary"`
	Trend           Trend           `json:"trend"`
	Alerts          []Alert         `json:"alerts"`
}

// WindowResult is the JSON response for window-summary endpoints (legacy + upgraded).
type WindowResult struct {
	Windows  []WindowSummary `json:"windows"`
	Warnings []Warning       `json:"warnings,omitempty"`
}

// SummaryResult is the JSON response for the summary endpoint.
type SummaryResult struct {
	Metric    string    `json:"metric"`
	Dimension string    `json:"dimension"`
	Summary   Summary   `json:"summary"`
	Warnings  []Warning `json:"warnings,omitempty"`
}

// ReportResult is the JSON response for the report endpoint.
type ReportResult struct {
	Report   Report    `json:"report"`
	Warnings []Warning `json:"warnings,omitempty"`
}

// QueryParams holds parsed query parameters for API requests.
type QueryParams struct {
	Metric           string
	Dimension        string
	WindowSize       time.Duration
	FillEmptyWindows bool
	GroupBy          string // "" | "metric" | "metric+dimension"
}
