package calc

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- ParseCSV ---

func TestParseCSVNormal(t *testing.T) {
	input := "timestamp,value\n2026-04-01T10:00:00Z,10\n2026-04-01T10:02:00Z,20\n"
	events, errs := ParseCSV(strings.NewReader(input))
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Value != 10 || events[1].Value != 20 {
		t.Fatalf("unexpected values: %v", events)
	}
}

func TestParseCSVEmptyFile(t *testing.T) {
	events, errs := ParseCSV(strings.NewReader(""))
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestParseCSVHeaderOnly(t *testing.T) {
	events, errs := ParseCSV(strings.NewReader("timestamp,value\n"))
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestParseCSVInvalidHeader(t *testing.T) {
	events, errs := ParseCSV(strings.NewReader("time,val\n1,2\n"))
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "invalid CSV header") {
		t.Fatalf("expected header error, got %v", errs)
	}
}

func TestParseCSVInvalidTimestamp(t *testing.T) {
	input := "timestamp,value\nnot-a-time,10\n2026-04-01T10:00:00Z,20\n"
	events, errs := ParseCSV(strings.NewReader(input))
	if len(events) != 1 {
		t.Fatalf("expected 1 valid event, got %d", len(events))
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "invalid timestamp") {
		t.Fatalf("expected timestamp error, got %v", errs)
	}
}

func TestParseCSVInvalidValue(t *testing.T) {
	input := "timestamp,value\n2026-04-01T10:00:00Z,abc\n2026-04-01T10:01:00Z,5\n"
	events, errs := ParseCSV(strings.NewReader(input))
	if len(events) != 1 {
		t.Fatalf("expected 1 valid event, got %d", len(events))
	}
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "invalid value") {
		t.Fatalf("expected value error, got %v", errs)
	}
}

func TestParseCSVMissingColumns(t *testing.T) {
	input := "timestamp,value\n2026-04-01T10:00:00Z\n"
	events, errs := ParseCSV(strings.NewReader(input))
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

// --- BuildWindowSummaries ---

func TestBuildWindowSummariesEmpty(t *testing.T) {
	result := BuildWindowSummaries(nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestBuildWindowSummariesSingleWindow(t *testing.T) {
	events := []Event{
		{Timestamp: mustParseTime("2026-04-01T10:00:00Z"), Value: 10},
		{Timestamp: mustParseTime("2026-04-01T10:02:00Z"), Value: 20},
		{Timestamp: mustParseTime("2026-04-01T10:04:59Z"), Value: 30},
	}
	windows := BuildWindowSummaries(events)
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
	w := windows[0]
	if w.WindowStart != "2026-04-01T10:00:00Z" {
		t.Fatalf("unexpected window_start: %s", w.WindowStart)
	}
	if w.WindowEnd != "2026-04-01T10:05:00Z" {
		t.Fatalf("unexpected window_end: %s", w.WindowEnd)
	}
	if w.Summary.Count != 3 {
		t.Fatalf("expected count 3, got %d", w.Summary.Count)
	}
	if w.Summary.Sum != 60 {
		t.Fatalf("expected sum 60, got %f", w.Summary.Sum)
	}
	if w.Summary.Min != 10 {
		t.Fatalf("expected min 10, got %f", w.Summary.Min)
	}
	if w.Summary.Max != 30 {
		t.Fatalf("expected max 30, got %f", w.Summary.Max)
	}
}

func TestBuildWindowSummariesMultipleWindows(t *testing.T) {
	events := []Event{
		{Timestamp: mustParseTime("2026-04-01T10:00:00Z"), Value: 10},
		{Timestamp: mustParseTime("2026-04-01T10:02:00Z"), Value: 20},
		{Timestamp: mustParseTime("2026-04-01T10:06:00Z"), Value: 30},
		{Timestamp: mustParseTime("2026-04-01T10:11:00Z"), Value: 40},
	}
	windows := BuildWindowSummaries(events)
	if len(windows) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(windows))
	}
	if windows[0].Summary.Count != 2 {
		t.Fatalf("window 0: expected count 2, got %d", windows[0].Summary.Count)
	}
	if windows[1].Summary.Count != 1 {
		t.Fatalf("window 1: expected count 1, got %d", windows[1].Summary.Count)
	}
	if windows[2].Summary.Count != 1 {
		t.Fatalf("window 2: expected count 1, got %d", windows[2].Summary.Count)
	}
}

func TestBuildWindowSummariesOutOfOrder(t *testing.T) {
	events := []Event{
		{Timestamp: mustParseTime("2026-04-01T10:06:00Z"), Value: 30},
		{Timestamp: mustParseTime("2026-04-01T10:00:00Z"), Value: 10},
		{Timestamp: mustParseTime("2026-04-01T10:02:00Z"), Value: 20},
	}
	windows := BuildWindowSummaries(events)
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(windows))
	}
	if windows[0].WindowStart != "2026-04-01T10:00:00Z" {
		t.Fatalf("first window should start at 10:00, got %s", windows[0].WindowStart)
	}
	if windows[0].Summary.Count != 2 {
		t.Fatalf("first window: expected count 2, got %d", windows[0].Summary.Count)
	}
}

func TestBuildWindowSummariesBoundaryPoint(t *testing.T) {
	// An event exactly at a window boundary belongs to the new window.
	events := []Event{
		{Timestamp: mustParseTime("2026-04-01T10:04:59Z"), Value: 1},
		{Timestamp: mustParseTime("2026-04-01T10:05:00Z"), Value: 2},
	}
	windows := BuildWindowSummaries(events)
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(windows))
	}
	if windows[0].Summary.Count != 1 || windows[0].Summary.Sum != 1 {
		t.Fatalf("window 0: expected count=1 sum=1, got count=%d sum=%f", windows[0].Summary.Count, windows[0].Summary.Sum)
	}
	if windows[1].Summary.Count != 1 || windows[1].Summary.Sum != 2 {
		t.Fatalf("window 1: expected count=1 sum=2, got count=%d sum=%f", windows[1].Summary.Count, windows[1].Summary.Sum)
	}
}

// --- HTTP Handler ---

func TestHandleWindowSummarySuccess(t *testing.T) {
	csvData := "timestamp,value\n2026-04-01T10:00:00Z,10\n2026-04-01T10:02:00Z,20\n"
	body, contentType := createMultipartForm(t, "file", "data.csv", csvData)

	req := httptest.NewRequest(http.MethodPost, "/api/window-summary", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	HandleWindowSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result WindowResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(result.Windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(result.Windows))
	}
}

func TestHandleWindowSummaryEmptyCSV(t *testing.T) {
	csvData := "timestamp,value\n"
	body, contentType := createMultipartForm(t, "file", "data.csv", csvData)

	req := httptest.NewRequest(http.MethodPost, "/api/window-summary", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	HandleWindowSummary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result WindowResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(result.Windows) != 0 {
		t.Fatalf("expected 0 windows, got %d", len(result.Windows))
	}
}

func TestHandleWindowSummaryMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/window-summary", nil)
	rr := httptest.NewRecorder()

	HandleWindowSummary(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHandleWindowSummaryBadCSV(t *testing.T) {
	csvData := "bad,header\n1,2\n"
	body, contentType := createMultipartForm(t, "file", "data.csv", csvData)

	req := httptest.NewRequest(http.MethodPost, "/api/window-summary", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	HandleWindowSummary(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- Helpers ---

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func createMultipartForm(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte(content))
	writer.Close()
	return &buf, writer.FormDataContentType()
}
