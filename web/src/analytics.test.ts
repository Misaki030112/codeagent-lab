/**
 * Tests for analytics TypeScript modules.
 * These tests verify type correctness, formatting, transforms, and query building.
 */

import { emptySummary } from "./analytics/types";
import type { Summary, Report, WindowSummary, Alert } from "./analytics/types";
import {
  formatNumber,
  formatPercent,
  formatDelta,
  trendLabel,
  trendClass,
  alertLabel,
  alertClass,
  formatWindowSize,
  formatTimestamp,
  formatSummaryField,
} from "./analytics/format";
import {
  summaryToTableRows,
  windowsToChartData,
  reportToNavItems,
  groupAlertsByType,
  compareReports,
} from "./analytics/transform";
import {
  buildQueryString,
  parseQueryString,
  buildAPIUrl,
  buildSummaryUrl,
  buildReportUrl,
  paramsToHash,
  paramsFromHash,
} from "./analytics/query";

// Also test existing string utilities
import {
  slugify,
  formatMetricLabel,
  buildSummaryAnchor,
  truncateMiddle,
} from "./string";

// ---------------------------------------------------------------------------
// Type tests
// ---------------------------------------------------------------------------

describe("emptySummary", () => {
  it("returns a zero-valued summary", () => {
    const s = emptySummary();
    expect(s.count).toBe(0);
    expect(s.sum).toBe(0);
    expect(s.min).toBe(0);
    expect(s.max).toBe(0);
    expect(s.average).toBe(0);
    expect(s.median).toBe(0);
    expect(s.variance).toBe(0);
    expect(s.std_dev).toBe(0);
    expect(s.p90).toBe(0);
    expect(s.p95).toBe(0);
    expect(s.first).toBe(0);
    expect(s.last).toBe(0);
    expect(s.delta).toBe(0);
    expect(s.percent_change).toBeNull();
  });

  it("has all 14 required fields", () => {
    const s = emptySummary();
    const keys = Object.keys(s);
    expect(keys).toHaveLength(14);
    expect(keys).toContain("count");
    expect(keys).toContain("variance");
    expect(keys).toContain("percent_change");
  });
});

// ---------------------------------------------------------------------------
// Format tests
// ---------------------------------------------------------------------------

describe("formatNumber", () => {
  it("formats a number with decimals", () => {
    expect(formatNumber(123.456)).toBe("123.46");
    expect(formatNumber(123.456, 1)).toBe("123.5");
  });

  it("returns — for null", () => {
    expect(formatNumber(null)).toBe("—");
    expect(formatNumber(undefined)).toBe("—");
  });
});

describe("formatPercent", () => {
  it("formats positive with + sign", () => {
    expect(formatPercent(25.5)).toBe("+25.5%");
  });

  it("formats negative without + sign", () => {
    expect(formatPercent(-10.2)).toBe("-10.2%");
  });

  it("returns — for null", () => {
    expect(formatPercent(null)).toBe("—");
  });
});

describe("formatDelta", () => {
  it("formats positive with + sign", () => {
    expect(formatDelta(5.0)).toBe("+5.00");
  });

  it("formats negative without + sign", () => {
    expect(formatDelta(-3.5)).toBe("-3.50");
  });
});

describe("trendLabel", () => {
  it("maps all trend values", () => {
    expect(trendLabel("up")).toBe("Trending Up");
    expect(trendLabel("down")).toBe("Trending Down");
    expect(trendLabel("flat")).toBe("Flat");
    expect(trendLabel("insufficient_data")).toBe("Insufficient Data");
  });
});

describe("trendClass", () => {
  it("produces CSS class names", () => {
    expect(trendClass("up")).toBe("trend-up");
    expect(trendClass("insufficient_data")).toBe("trend-insufficient-data");
  });
});

describe("alertLabel", () => {
  it("maps all alert types", () => {
    expect(alertLabel("spike")).toBe("Spike Detected");
    expect(alertLabel("drop")).toBe("Drop Detected");
    expect(alertLabel("high_variance")).toBe("High Variance");
  });
});

describe("alertClass", () => {
  it("produces CSS class names", () => {
    expect(alertClass("high_variance")).toBe("alert-high-variance");
  });
});

describe("formatWindowSize", () => {
  it("formats minutes", () => {
    expect(formatWindowSize("5m")).toBe("5 min");
    expect(formatWindowSize("1m")).toBe("1 min");
  });

  it("formats hours", () => {
    expect(formatWindowSize("1h")).toBe("1 hour");
    expect(formatWindowSize("2h")).toBe("2 hours");
  });
});

describe("formatTimestamp", () => {
  it("formats RFC3339 to readable", () => {
    const result = formatTimestamp("2026-04-01T10:00:00Z");
    expect(result).toContain("2026-04-01");
    expect(result).toContain("10:00:00");
  });

  it("returns input for invalid timestamps", () => {
    expect(formatTimestamp("not-a-date")).toBe("not-a-date");
  });
});

describe("formatSummaryField", () => {
  it("formats count as integer", () => {
    expect(formatSummaryField("count", 5)).toBe("5");
  });

  it("formats percent_change as percentage", () => {
    expect(formatSummaryField("percent_change", 25.5)).toBe("+25.5%");
  });

  it("formats delta with sign", () => {
    expect(formatSummaryField("delta", -3.5)).toBe("-3.50");
  });
});

// ---------------------------------------------------------------------------
// Transform tests
// ---------------------------------------------------------------------------

const sampleSummary: Summary = {
  count: 5,
  sum: 15,
  min: 1,
  max: 5,
  average: 3,
  median: 3,
  variance: 2,
  std_dev: 1.41,
  p90: 4.6,
  p95: 4.8,
  first: 1,
  last: 5,
  delta: 4,
  percent_change: 400,
};

describe("summaryToTableRows", () => {
  it("produces 14 rows", () => {
    const rows = summaryToTableRows(sampleSummary);
    expect(rows).toHaveLength(14);
  });

  it("includes labels for all fields", () => {
    const rows = summaryToTableRows(sampleSummary);
    const labels = rows.map((r) => r.label);
    expect(labels).toContain("Count");
    expect(labels).toContain("Variance");
    expect(labels).toContain("% Change");
  });

  it("formats count as integer", () => {
    const rows = summaryToTableRows(sampleSummary);
    const countRow = rows.find((r) => r.key === "count")!;
    expect(countRow.value).toBe("5");
  });
});

describe("windowsToChartData", () => {
  const windows: WindowSummary[] = [
    { window_start: "2026-04-01T10:00:00Z", window_end: "2026-04-01T10:05:00Z", summary: sampleSummary },
    { window_start: "2026-04-01T10:05:00Z", window_end: "2026-04-01T10:10:00Z", summary: { ...sampleSummary, average: 5 } },
  ];

  it("extracts average by default", () => {
    const data = windowsToChartData(windows);
    expect(data).toHaveLength(2);
    expect(data[0].value).toBe(3);
    expect(data[1].value).toBe(5);
  });

  it("extracts custom field", () => {
    const data = windowsToChartData(windows, "sum");
    expect(data[0].value).toBe(15);
  });
});

describe("reportToNavItems", () => {
  const report: Report = {
    generated_at: "2026-04-01T10:00:00Z",
    metric: "revenue",
    dimension: "",
    window_size: "5m",
    current_windows: [
      { window_start: "2026-04-01T10:00:00Z", window_end: "2026-04-01T10:05:00Z", summary: sampleSummary },
    ],
    previous_windows: [],
    overall_summary: sampleSummary,
    trend: "up",
    alerts: [{ type: "spike", window_start: "2026-04-01T10:00:00Z", message: "test" }],
  };

  it("includes overall, windows, and alerts sections", () => {
    const items = reportToNavItems(report);
    expect(items[0].id).toBe("overall");
    expect(items[1].id).toBe("window-0");
    expect(items[2].id).toBe("alerts");
  });
});

describe("groupAlertsByType", () => {
  it("groups alerts by type", () => {
    const alerts: Alert[] = [
      { type: "spike", window_start: "t1", message: "a" },
      { type: "drop", window_start: "t2", message: "b" },
      { type: "spike", window_start: "t3", message: "c" },
    ];
    const groups = groupAlertsByType(alerts);
    expect(groups.get("spike")!).toHaveLength(2);
    expect(groups.get("drop")!).toHaveLength(1);
  });
});

describe("compareReports", () => {
  it("computes comparison with previous", () => {
    const current: Report = {
      generated_at: "",
      metric: "revenue",
      dimension: "",
      window_size: "5m",
      current_windows: [],
      previous_windows: [],
      overall_summary: { ...sampleSummary, average: 150 },
      trend: "up",
      alerts: [],
    };
    const previous: Report = {
      ...current,
      overall_summary: { ...sampleSummary, average: 100 },
    };
    const result = compareReports(current, previous);
    expect(result.averageChange).toBeCloseTo(50);
  });

  it("handles null previous", () => {
    const current: Report = {
      generated_at: "",
      metric: "revenue",
      dimension: "",
      window_size: "5m",
      current_windows: [],
      previous_windows: [],
      overall_summary: sampleSummary,
      trend: "up",
      alerts: [],
    };
    const result = compareReports(current, null);
    expect(result.previousCount).toBe(0);
    expect(result.averageChange).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// Query tests
// ---------------------------------------------------------------------------

describe("buildQueryString", () => {
  it("builds from params", () => {
    const qs = buildQueryString({ metric: "revenue", window_size: "5m" });
    expect(qs).toContain("metric=revenue");
    expect(qs).toContain("window_size=5m");
  });

  it("returns empty for no params", () => {
    expect(buildQueryString({})).toBe("");
  });

  it("includes fill_empty_windows only when true", () => {
    const qs = buildQueryString({ fill_empty_windows: true });
    expect(qs).toContain("fill_empty_windows=true");
    const qs2 = buildQueryString({ fill_empty_windows: false });
    expect(qs2).not.toContain("fill_empty_windows");
  });
});

describe("parseQueryString", () => {
  it("parses query string", () => {
    const params = parseQueryString("metric=revenue&window_size=5m&fill_empty_windows=true");
    expect(params.metric).toBe("revenue");
    expect(params.window_size).toBe("5m");
    expect(params.fill_empty_windows).toBe(true);
  });
});

describe("buildAPIUrl", () => {
  it("builds full URL", () => {
    const url = buildAPIUrl("/api/report", { metric: "revenue" }, "http://localhost:8080");
    expect(url).toBe("http://localhost:8080/api/report?metric=revenue");
  });

  it("handles no params", () => {
    const url = buildAPIUrl("/api/report", {});
    expect(url).toBe("http://localhost:8080/api/report");
  });
});

describe("buildSummaryUrl / buildReportUrl", () => {
  it("uses correct endpoints", () => {
    expect(buildSummaryUrl({ metric: "revenue" })).toContain("/api/summary");
    expect(buildReportUrl({ metric: "revenue" })).toContain("/api/report");
  });
});

describe("paramsToHash / paramsFromHash", () => {
  it("round-trips", () => {
    const original = { metric: "revenue", window_size: "15m" };
    const hash = paramsToHash(original);
    const parsed = paramsFromHash(hash);
    expect(parsed.metric).toBe("revenue");
    expect(parsed.window_size).toBe("15m");
  });
});

// ---------------------------------------------------------------------------
// Existing string utility tests
// ---------------------------------------------------------------------------

describe("slugify", () => {
  it("converts to slug", () => {
    expect(slugify("Hello World")).toBe("hello-world");
    expect(slugify("Hello__World!!")).toBe("hello-world");
  });
});

describe("formatMetricLabel", () => {
  it("converts snake_case", () => {
    expect(formatMetricLabel("total_revenue")).toBe("Total Revenue");
  });

  it("converts camelCase", () => {
    expect(formatMetricLabel("avgSessionTime")).toBe("Avg Session Time");
  });
});

describe("buildSummaryAnchor", () => {
  it("generates anchor", () => {
    expect(buildSummaryAnchor("revenue", "Rev")).toBe('<a href="#revenue">Rev</a>');
  });
});

describe("truncateMiddle", () => {
  it("truncates long strings", () => {
    expect(truncateMiddle("abcdefghij", 7)).toBe("ab...ij");
  });

  it("returns short strings unchanged", () => {
    expect(truncateMiddle("abc", 5)).toBe("abc");
  });
});
