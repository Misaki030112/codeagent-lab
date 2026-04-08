/**
 * Cross-language analytics summary toolkit — Web display helpers.
 *
 * Pure functions for transforming strings in analytics dashboard UIs.
 */

/**
 * Convert an arbitrary string to a URL-safe slug.
 *
 * Handles multiple spaces, underscores, punctuation, consecutive
 * separators, and leading/trailing separators.
 */
export function slugify(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .replace(/_/g, '-')
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s-]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

/**
 * Format a metric key (snake_case or camelCase) as a human-readable
 * title-case label suitable for dashboard headings.
 *
 * Examples:
 *   "total_revenue"  → "Total Revenue"
 *   "avgSessionTime" → "Avg Session Time"
 */
export function formatMetricLabel(key: string): string {
  return key
    .replace(/_/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

/**
 * Build an HTML anchor element for in-page navigation within an
 * analytics summary page.
 *
 * Example:
 *   buildSummaryAnchor("revenue", "Total Revenue")
 *   → '<a href="#revenue">Total Revenue</a>'
 */
export function buildSummaryAnchor(section: string, label: string): string {
  const slug = slugify(section)
  return `<a href="#${slug}">${label}</a>`
}

/**
 * Truncate a string in the middle, preserving the start and end, with
 * an ellipsis in the centre.
 *
 * If `input.length <= max`, the string is returned unchanged.
 * `max` must be at least 4 (to fit at least one char + "..." + one char).
 *
 * Example:
 *   truncateMiddle("abcdefghij", 7) → "ab...ij"
 */
export function truncateMiddle(input: string, max: number): string {
  if (max < 4) {
    max = 4
  }
  if (input.length <= max) {
    return input
  }
  const keep = max - 3
  const front = Math.ceil(keep / 2)
  const back = Math.floor(keep / 2)
  return input.slice(0, front) + '...' + input.slice(input.length - back)
}
