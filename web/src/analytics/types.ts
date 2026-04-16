/**
 * Shared TypeScript type definitions for the analytics report contract.
 * These types mirror the JSON schema in fixtures/schema/report_schema.json
 * and align with Go structs and Python dicts.
 */

/** Statistical summary of a set of values. */
export interface Summary {
  count: number;
  sum: number;
  min: number;
  max: number;
  average: number;
  median: number;
  variance: number;
  std_dev: number;
  p90: number;
  p95: number;
  first: number;
  last: number;
  delta: number;
  percent_change: number | null;
}

/** Aggregated summary for a time window. */
export interface WindowSummary {
  window_start: string;
  window_end: string;
  summary: Summary;
}

/** Overall trend direction. */
export type Trend = "up" | "down" | "flat" | "insufficient_data";

/** Type of anomaly detected. */
export type AlertType = "spike" | "drop" | "high_variance";

/** A detected anomaly in a window. */
export interface Alert {
  type: AlertType;
  window_start: string;
  message: string;
  value?: number;
  threshold?: number;
}

/** Top-level analytics report. */
export interface Report {
  generated_at: string;
  metric: string;
  dimension: string;
  window_size: string;
  current_windows: WindowSummary[];
  previous_windows: WindowSummary[];
  overall_summary: Summary;
  trend: Trend;
  alerts: Alert[];
}

/** Structured parse warning. */
export interface Warning {
  row: number;
  message: string;
}

/** Unified API error response. */
export interface APIError {
  error: string;
  message: string;
}

/** Response from POST /api/summary. */
export interface SummaryResult {
  metric: string;
  dimension: string;
  summary: Summary;
  warnings?: Warning[];
}

/** Response from POST /api/window-summary (upgraded). */
export interface WindowResult {
  windows: WindowSummary[];
  warnings?: Warning[];
}

/** Response from POST /api/report. */
export interface ReportResult {
  report: Report;
  warnings?: Warning[];
}

/** Query parameters for analytics API requests. */
export interface QueryParams {
  metric?: string;
  dimension?: string;
  window_size?: string;
  fill_empty_windows?: boolean;
  group_by?: string;
}

/** A zero-valued Summary for empty inputs. */
export function emptySummary(): Summary {
  return {
    count: 0,
    sum: 0,
    min: 0,
    max: 0,
    average: 0,
    median: 0,
    variance: 0,
    std_dev: 0,
    p90: 0,
    p95: 0,
    first: 0,
    last: 0,
    delta: 0,
    percent_change: null,
  };
}
