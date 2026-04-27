package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	calc "codeagent-lab/calc"
)

const version = "2.0.0"

// maxUploadSize is the maximum allowed request body size (10 MB).
const maxUploadSize = 10 << 20

func main() {
	mux := http.NewServeMux()

	// Legacy endpoint (backward compatible)
	mux.HandleFunc("/api/window-summary", calc.HandleWindowSummary)

	// New analytics endpoints
	mux.HandleFunc("/api/summary", handleSummary)
	mux.HandleFunc("/api/report", handleReport)
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/api/meta", handleMeta)

	addr := ":8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	fmt.Printf("Listening on %s\n", addr)
	log.Fatal(srv.ListenAndServe())
}

// readUploadedCSV is a shared helper that validates the HTTP method, reads the
// uploaded file, and parses it as multi-metric CSV. It writes an error response
// and returns nil if anything goes wrong.
func readUploadedCSV(w http.ResponseWriter, r *http.Request) ([]calc.MultiEvent, []calc.Warning, multipart.File) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
		return nil, nil, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("failed to read uploaded file: %v", err))
		return nil, nil, nil
	}

	events, warnings := calc.ParseMultiCSV(file)
	if len(events) == 0 && len(warnings) > 0 {
		file.Close()
		writeError(w, http.StatusBadRequest, "parse_error", "CSV parsing failed")
		return nil, nil, nil
	}

	return events, warnings, file
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	events, warnings, file := readUploadedCSV(w, r)
	if file == nil {
		return
	}
	defer file.Close()

	metric := r.FormValue("metric")
	dimension := r.FormValue("dimension")

	if metric == "" && len(events) > 0 {
		metric = events[0].Metric
	}

	filtered := calc.FilterEvents(events, metric, dimension)
	values := make([]float64, len(filtered))
	for i, e := range filtered {
		values[i] = e.Value
	}

	writeJSON(w, calc.SummaryResult{
		Metric:    metric,
		Dimension: dimension,
		Summary:   calc.BuildSummary(values),
		Warnings:  warnings,
	})
}

func handleReport(w http.ResponseWriter, r *http.Request) {
	events, warnings, file := readUploadedCSV(w, r)
	if file == nil {
		return
	}
	defer file.Close()

	metric := r.FormValue("metric")
	dimension := r.FormValue("dimension")
	windowSizeStr := r.FormValue("window_size")
	fillEmpty := r.FormValue("fill_empty_windows") == "true"

	if metric == "" && len(events) > 0 {
		metric = events[0].Metric
	}

	ws, err := calc.ParseWindowSize(windowSizeStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_param", err.Error())
		return
	}

	writeJSON(w, calc.ReportResult{
		Report:   calc.BuildReport(events, metric, dimension, ws, fillEmpty),
		Warnings: warnings,
	})
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{
		"version": version,
		"name":    "codeagent-lab analytics API",
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(calc.APIError{Error: errCode, Message: message})
}
