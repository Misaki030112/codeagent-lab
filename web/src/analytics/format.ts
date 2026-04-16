/**
 * Formatting utilities for analytics values, labels, and trends.
 * Designed for dashboard display of summary/report data.
 */

import type { Trend, AlertType, Summary } from "./types";

/**
 * Format a number with fixed decimal places. Returns "—" for null/undefined.
 */
export function formatNumber(value: number | null | undefined, decimals = 2): string {
  if (value === null || value === undefined) {
    return "—";
  }
  return value.toFixed(decimals);
}

/**
 * Format a percentage value with a "%" suffix. Returns "—" for null.
 */
export function formatPercent(value: number | null | undefined, decimals = 1): string {
  if (value === null || value === undefined) {
    return "—";
  }
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(decimals)}%`;
}

/**
 * Format a delta value with a sign prefix.
 */
export function formatDelta(value: number, decimals = 2): string {
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(decimals)}`;
}

/**
 * Return a human-readable label for a trend value.
 */
export function trendLabel(trend: Trend): string {
  switch (trend) {
    case "up":
      return "Trending Up";
    case "down":
      return "Trending Down";
    case "flat":
      return "Flat";
    case "insufficient_data":
      return "Insufficient Data";
  }
}

/**
 * Return a CSS-friendly class name for a trend value.
 */
export function trendClass(trend: Trend): string {
  return `trend-${trend.replace("_", "-")}`;
}

/**
 * Return a human-readable label for an alert type.
 */
export function alertLabel(alertType: AlertType): string {
  switch (alertType) {
    case "spike":
      return "Spike Detected";
    case "drop":
      return "Drop Detected";
    case "high_variance":
      return "High Variance";
  }
}

/**
 * Return a CSS-friendly class name for an alert type.
 */
export function alertClass(alertType: AlertType): string {
  return `alert-${alertType.replace("_", "-")}`;
}

/**
 * Format a window size string for display (e.g., "5m" -> "5 min", "1h" -> "1 hour").
 */
export function formatWindowSize(ws: string): string {
  const match = ws.match(/^(\d+)(m|h)$/);
  if (!match) return ws;
  const num = parseInt(match[1], 10);
  const unit = match[2];
  if (unit === "h") {
    return num === 1 ? "1 hour" : `${num} hours`;
  }
  return num === 1 ? "1 min" : `${num} min`;
}

/**
 * Format an RFC3339 timestamp for display.
 */
export function formatTimestamp(ts: string): string {
  const d = new Date(ts);
  if (isNaN(d.getTime())) return ts;
  return d.toISOString().replace("T", " ").replace(/\.000Z$/, " UTC");
}

/**
 * Format a summary field for display, choosing appropriate formatting
 * based on the field name.
 */
export function formatSummaryField(key: keyof Summary, value: number | null): string {
  if (key === "count") {
    return value !== null ? value.toString() : "0";
  }
  if (key === "percent_change") {
    return formatPercent(value);
  }
  if (key === "delta") {
    return value !== null ? formatDelta(value) : "—";
  }
  return formatNumber(value);
}
