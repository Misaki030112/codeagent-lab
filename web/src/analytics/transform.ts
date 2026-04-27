/**
 * Transform functions to convert Report data into view-model structures
 * suitable for tables, charts, and navigation.
 */

import type { Report, Summary, WindowSummary, Alert } from "./types";
import { formatSummaryField } from "./format";

/** A row in a summary table. */
export interface SummaryTableRow {
  label: string;
  value: string;
  key: keyof Summary;
  raw: number | null;
}

/** A data point for chart rendering. */
export interface ChartDataPoint {
  label: string;
  windowStart: string;
  windowEnd: string;
  value: number;
}

/** A navigation anchor item. */
export interface NavItem {
  id: string;
  label: string;
  href: string;
}

/** The summary field labels in display order. */
const SUMMARY_FIELD_LABELS: Array<{ key: keyof Summary; label: string }> = [
  { key: "count", label: "Count" },
  { key: "sum", label: "Sum" },
  { key: "min", label: "Min" },
  { key: "max", label: "Max" },
  { key: "average", label: "Average" },
  { key: "median", label: "Median" },
  { key: "variance", label: "Variance" },
  { key: "std_dev", label: "Std Dev" },
  { key: "p90", label: "P90" },
  { key: "p95", label: "P95" },
  { key: "first", label: "First" },
  { key: "last", label: "Last" },
  { key: "delta", label: "Delta" },
  { key: "percent_change", label: "% Change" },
];

/**
 * Convert a Summary into table rows for display.
 */
export function summaryToTableRows(summary: Summary): SummaryTableRow[] {
  return SUMMARY_FIELD_LABELS.map(({ key, label }) => {
    const raw = summary[key];
    const value = formatSummaryField(key, raw as number | null);
    return { label, value, key, raw: raw as number | null };
  });
}

/**
 * Extract chart data points from window summaries.
 * Uses the specified summary field as the y-axis value.
 */
export function windowsToChartData(
  windows: WindowSummary[],
  field: keyof Summary = "average",
): ChartDataPoint[] {
  return windows.map((w) => {
    const value = w.summary[field];
    return {
      label: w.window_start,
      windowStart: w.window_start,
      windowEnd: w.window_end,
      value: typeof value === "number" ? value : 0,
    };
  });
}

/**
 * Generate navigation items from a report for building a table-of-contents.
 */
export function reportToNavItems(report: Report): NavItem[] {
  const items: NavItem[] = [
    { id: "overall", label: "Overall Summary", href: "#overall" },
  ];

  report.current_windows.forEach((w, i) => {
    const id = `window-${i}`;
    items.push({
      id,
      label: `Window ${i + 1}: ${w.window_start}`,
      href: `#${id}`,
    });
  });

  if (report.alerts.length > 0) {
    items.push({ id: "alerts", label: "Alerts", href: "#alerts" });
  }

  return items;
}

/**
 * Group alerts by their type for sectioned display.
 */
export function groupAlertsByType(alerts: Alert[]): Map<string, Alert[]> {
  const groups = new Map<string, Alert[]>();
  for (const alert of alerts) {
    const existing = groups.get(alert.type) || [];
    existing.push(alert);
    groups.set(alert.type, existing);
  }
  return groups;
}

/**
 * Compare two reports and produce a comparison summary.
 */
export interface ComparisonResult {
  metric: string;
  currentCount: number;
  previousCount: number;
  currentAverage: number;
  previousAverage: number;
  averageChange: number | null;
  trend: string;
}

export function compareReports(
  current: Report,
  previous: Report | null,
): ComparisonResult {
  const currentAvg = current.overall_summary.average;
  const previousAvg = previous?.overall_summary.average ?? 0;
  const previousCount = previous?.overall_summary.count ?? 0;

  let averageChange: number | null = null;
  if (previous && previousAvg !== 0) {
    averageChange = ((currentAvg - previousAvg) / Math.abs(previousAvg)) * 100;
  }

  return {
    metric: current.metric,
    currentCount: current.overall_summary.count,
    previousCount,
    currentAverage: currentAvg,
    previousAverage: previousAvg,
    averageChange,
    trend: current.trend,
  };
}
