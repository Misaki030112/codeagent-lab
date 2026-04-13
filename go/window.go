package calc

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// WindowSize is the fixed duration for each aggregation bucket.
const WindowSize = 5 * time.Minute

// Event represents a single timestamped numeric observation.
type Event struct {
	Timestamp time.Time
	Value     float64
}

// WindowSummary holds the time bounds and statistics for one window.
type WindowSummary struct {
	WindowStart string  `json:"window_start"`
	WindowEnd   string  `json:"window_end"`
	Summary     Summary `json:"summary"`
}

// WindowResult is the top-level JSON response containing all windows.
type WindowResult struct {
	Windows  []WindowSummary `json:"windows"`
	Warnings []string        `json:"warnings,omitempty"`
}

// maxUploadSize is the maximum allowed request body size for CSV uploads (10 MB).
const maxUploadSize = 10 << 20

// ParseCSV reads timestamp,value rows from a CSV reader.
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

// windowStart returns the start of the 5-minute bucket containing t.
func windowStart(t time.Time) time.Time {
	return t.UTC().Truncate(WindowSize)
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
		ws := windowStart(e.Timestamp)
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
	enc.Encode(WindowResult{Windows: windows, Warnings: warnings})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
