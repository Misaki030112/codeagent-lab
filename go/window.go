package calc

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DefaultWindowSize is the default duration for each aggregation bucket.
const DefaultWindowSize = 5 * time.Minute

// WindowSize is kept for backward compatibility.
const WindowSize = DefaultWindowSize

// maxUploadSize is the maximum allowed request body size for CSV uploads (10 MB).
const maxUploadSize = 10 << 20

// Event represents a single timestamped numeric observation (legacy format).
type Event struct {
	Timestamp time.Time
	Value     float64
}

// ParseWindowSize parses a window size string like "1m", "5m", "15m", "1h".
// Returns DefaultWindowSize for empty string.
func ParseWindowSize(s string) (time.Duration, error) {
	if s == "" {
		return DefaultWindowSize, nil
	}
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid window size: %q", s)
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("invalid window size: %q", s)
	}
	switch unit {
	case 'm':
		return time.Duration(num) * time.Minute, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid window size unit %q: expected 'm' or 'h'", string(unit))
	}
}

// FormatWindowSize formats a duration as a human-readable window size string.
func FormatWindowSize(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// ParseCSV reads timestamp,value rows from a CSV reader (legacy format).
// It expects a header row with columns "timestamp" and "value".
// Invalid rows are collected as errors but do not stop parsing.
func ParseCSV(r io.Reader) ([]Event, []error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read CSV header: %w", err)}
	}
	if len(header) < 2 || header[0] != "timestamp" || header[1] != "value" {
		return nil, []error{fmt.Errorf("invalid CSV header: expected [timestamp, value], got %v", header)}
	}

	var events []Event
	var errs []error
	lineNum := 1

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", lineNum, err))
			continue
		}
		if len(record) < 2 {
			errs = append(errs, fmt.Errorf("line %d: expected 2 columns, got %d", lineNum, len(record)))
			continue
		}

		ts, err := time.Parse(time.RFC3339, record[0])
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: invalid timestamp %q: %w", lineNum, record[0], err))
			continue
		}

		val, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: invalid value %q: %w", lineNum, record[1], err))
			continue
		}

		events = append(events, Event{Timestamp: ts, Value: val})
	}

	return events, errs
}

// ParseMultiCSV reads multi-metric CSV rows from a reader.
// Expected header: timestamp,metric,value,dimension,source
// metric is required; dimension and source may be empty.
// Invalid rows produce warnings but do not stop parsing.
func ParseMultiCSV(r io.Reader) ([]MultiEvent, []Warning) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		return nil, []Warning{{Row: 1, Message: fmt.Sprintf("failed to read CSV header: %v", err)}}
	}
	if len(header) < 3 || header[0] != "timestamp" || header[1] != "metric" || header[2] != "value" {
		return nil, []Warning{{Row: 1, Message: fmt.Sprintf("invalid CSV header: expected [timestamp,metric,value,...], got %v", header)}}
	}

	var events []MultiEvent
	var warnings []Warning
	lineNum := 1

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			warnings = append(warnings, Warning{Row: lineNum, Message: err.Error()})
			continue
		}
		if len(record) < 3 {
			warnings = append(warnings, Warning{Row: lineNum, Message: fmt.Sprintf("expected at least 3 columns, got %d", len(record))})
			continue
		}

		ts, err := time.Parse(time.RFC3339, record[0])
		if err != nil {
			warnings = append(warnings, Warning{Row: lineNum, Message: fmt.Sprintf("invalid timestamp %q: %v", record[0], err)})
			continue
		}

		metric := strings.TrimSpace(record[1])
		if metric == "" {
			warnings = append(warnings, Warning{Row: lineNum, Message: "empty metric field"})
			continue
		}

		val, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			warnings = append(warnings, Warning{Row: lineNum, Message: fmt.Sprintf("invalid value %q: %v", record[2], err)})
			continue
		}

		evt := MultiEvent{
			Timestamp: ts,
			Metric:    metric,
			Value:     val,
		}
		if len(record) > 3 {
			evt.Dimension = strings.TrimSpace(record[3])
		}
		if len(record) > 4 {
			evt.Source = strings.TrimSpace(record[4])
		}

		events = append(events, evt)
	}

	return events, warnings
}

// FilterEvents filters multi-events by metric and optional dimension.
func FilterEvents(events []MultiEvent, metric, dimension string) []MultiEvent {
	var result []MultiEvent
	for _, e := range events {
		if e.Metric != metric {
			continue
		}
		if dimension != "" && e.Dimension != dimension {
			continue
		}
		result = append(result, e)
	}
	return result
}

// windowStartFor returns the start of the bucket containing t for the given window size.
func windowStartFor(t time.Time, ws time.Duration) time.Time {
	utc := t.UTC()
	epoch := utc.Unix()
	wsSec := int64(ws.Seconds())
	bucketStart := epoch - (epoch % wsSec)
	return time.Unix(bucketStart, 0).UTC()
}

// BuildWindowSummaries groups events into fixed 5-minute windows using
// [start, end) semantics and computes a Summary for each window.
// Out-of-order input is handled gracefully. Empty windows are not emitted.
func BuildWindowSummaries(events []Event) []WindowSummary {
	if len(events) == 0 {
		return nil
	}

	buckets := make(map[time.Time][]float64)
	for _, e := range events {
		ws := windowStartFor(e.Timestamp, WindowSize)
		buckets[ws] = append(buckets[ws], e.Value)
	}

	starts := make([]time.Time, 0, len(buckets))
	for ws := range buckets {
		starts = append(starts, ws)
	}
	sort.Slice(starts, func(i, j int) bool {
		return starts[i].Before(starts[j])
	})

	result := make([]WindowSummary, 0, len(starts))
	for _, ws := range starts {
		result = append(result, WindowSummary{
			WindowStart: ws.Format(time.RFC3339),
			WindowEnd:   ws.Add(WindowSize).Format(time.RFC3339),
			Summary:     BuildSummary(buckets[ws]),
		})
	}

	return result
}

// multiEventBucket tracks values in a time bucket.
// Events are pre-sorted by timestamp before bucketing, so insertion order
// is already chronological.
type multiEventBucket struct {
	values []float64
}

// BuildMultiWindowSummaries groups multi-events into configurable windows.
// If fillEmpty is true, empty windows between min and max timestamps are filled with zero summaries.
func BuildMultiWindowSummaries(events []MultiEvent, ws time.Duration, fillEmpty bool) []WindowSummary {
	if len(events) == 0 {
		return nil
	}

	// Sort by timestamp for stable ordering
	sorted := make([]MultiEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	buckets := make(map[time.Time]*multiEventBucket)
	for _, e := range sorted {
		start := windowStartFor(e.Timestamp, ws)
		b, ok := buckets[start]
		if !ok {
			b = &multiEventBucket{}
			buckets[start] = b
		}
		b.values = append(b.values, e.Value)
	}

	// Collect and sort bucket starts
	starts := make([]time.Time, 0, len(buckets))
	for s := range buckets {
		starts = append(starts, s)
	}
	sort.Slice(starts, func(i, j int) bool {
		return starts[i].Before(starts[j])
	})

	// If filling empty windows, create entries for all windows in range
	if fillEmpty && len(starts) > 1 {
		minStart := starts[0]
		maxStart := starts[len(starts)-1]
		allStarts := make([]time.Time, 0)
		for t := minStart; !t.After(maxStart); t = t.Add(ws) {
			allStarts = append(allStarts, t)
			if _, ok := buckets[t]; !ok {
				buckets[t] = &multiEventBucket{}
			}
		}
		starts = allStarts
	}

	result := make([]WindowSummary, 0, len(starts))
	for _, s := range starts {
		b := buckets[s]
		result = append(result, WindowSummary{
			WindowStart: s.Format(time.RFC3339),
			WindowEnd:   s.Add(ws).Format(time.RFC3339),
			Summary:     BuildSummaryOrdered(b.values, b.values),
		})
	}

	return result
}

// HandleWindowSummary is an HTTP handler that accepts a CSV file via
// POST multipart form upload (field name "file") and returns the
// per-window summary JSON.
func HandleWindowSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "only POST is allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to read uploaded file: %v", err))
		return
	}
	defer file.Close()

	events, parseErrs := ParseCSV(file)
	if len(events) == 0 && len(parseErrs) > 0 {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("CSV parsing failed: %v", parseErrs))
		return
	}

	windows := BuildWindowSummaries(events)
	if windows == nil {
		windows = []WindowSummary{}
	}

	var warnings []string
	for _, e := range parseErrs {
		warnings = append(warnings, e.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(struct {
		Windows  []WindowSummary `json:"windows"`
		Warnings []string        `json:"warnings,omitempty"`
	}{Windows: windows, Warnings: warnings})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Legacy error format: single "error" field for backward compatibility.
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
