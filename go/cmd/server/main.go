package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	calc "codeagent-lab/calc"
)

const version = "2.0.0"

func main() {
	// Legacy endpoint (backward compatible)
	http.HandleFunc("/api/window-summary", calc.HandleWindowSummary)

	// New analytics endpoints
	http.HandleFunc("/api/summary", handleSummary)
	http.HandleFunc("/api/report", handleReport)
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/api/meta", handleMeta)

	addr := ":8080"
	fmt.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("failed to read uploaded file: %v", err))
		return
	}
	defer file.Close()

	events, warnings := calc.ParseMultiCSV(file)
	if len(events) == 0 && len(warnings) > 0 {
		writeError(w, http.StatusBadRequest, "parse_error", "CSV parsing failed")
		return
	}

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

	summary := calc.BuildSummary(values)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(calc.SummaryResult{
		Metric:    metric,
		Dimension: dimension,
		Summary:   summary,
		Warnings:  warnings,
	})
}

func handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("failed to read uploaded file: %v", err))
		return
	}
	defer file.Close()

	events, warnings := calc.ParseMultiCSV(file)
	if len(events) == 0 && len(warnings) > 0 {
		writeError(w, http.StatusBadRequest, "parse_error", "CSV parsing failed")
		return
	}

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

	report := calc.BuildReport(events, metric, dimension, ws, fillEmpty)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(calc.ReportResult{
		Report:   report,
		Warnings: warnings,
	})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version,
		"name":    "codeagent-lab analytics API",
	})
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(calc.APIError{Error: errCode, Message: message})
}
