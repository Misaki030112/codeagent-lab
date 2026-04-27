package calc

// APIError is the unified JSON error response for all API endpoints.
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Warning represents a structured parse warning tied to a specific CSV row.
type Warning struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}
