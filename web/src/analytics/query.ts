/**
 * Query builder and URL state utilities for the analytics API.
 */

import type { QueryParams } from "./types";

/** Default API base URL. */
const DEFAULT_BASE_URL = "http://localhost:8080";

/**
 * Build a URL search string from query params.
 */
export function buildQueryString(params: QueryParams): string {
  const sp = new URLSearchParams();
  if (params.metric) sp.set("metric", params.metric);
  if (params.dimension) sp.set("dimension", params.dimension);
  if (params.window_size) sp.set("window_size", params.window_size);
  if (params.fill_empty_windows) sp.set("fill_empty_windows", "true");
  if (params.group_by) sp.set("group_by", params.group_by);
  return sp.toString();
}

/**
 * Parse query params from a URL search string.
 */
export function parseQueryString(search: string): QueryParams {
  const params = new URLSearchParams(search);
  return {
    metric: params.get("metric") || undefined,
    dimension: params.get("dimension") || undefined,
    window_size: params.get("window_size") || undefined,
    fill_empty_windows: params.get("fill_empty_windows") === "true",
    group_by: params.get("group_by") || undefined,
  };
}

/**
 * Build the full URL for an API endpoint with query params.
 */
export function buildAPIUrl(
  endpoint: string,
  params: QueryParams,
  baseUrl = DEFAULT_BASE_URL,
): string {
  const qs = buildQueryString(params);
  const sep = qs ? "?" : "";
  return `${baseUrl}${endpoint}${sep}${qs}`;
}

/**
 * Build a summary API URL.
 */
export function buildSummaryUrl(params: QueryParams, baseUrl?: string): string {
  return buildAPIUrl("/api/summary", params, baseUrl);
}

/**
 * Build a window-summary API URL.
 */
export function buildWindowSummaryUrl(params: QueryParams, baseUrl?: string): string {
  return buildAPIUrl("/api/window-summary", params, baseUrl);
}

/**
 * Build a report API URL.
 */
export function buildReportUrl(params: QueryParams, baseUrl?: string): string {
  return buildAPIUrl("/api/report", params, baseUrl);
}

/**
 * Serialize query params to a URL hash for browser state.
 */
export function paramsToHash(params: QueryParams): string {
  return "#" + buildQueryString(params);
}

/**
 * Parse query params from a URL hash.
 */
export function paramsFromHash(hash: string): QueryParams {
  const search = hash.startsWith("#") ? hash.slice(1) : hash;
  return parseQueryString(search);
}
