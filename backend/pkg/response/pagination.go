package response

// PaginatedResponse represents a paginated list response
type PaginatedResponse struct {
	Items      any    `json:"items"`
	Total      int    `json:"total"`
	Continue   string `json:"continue,omitempty"`
	Remaining  int64  `json:"remainingItemCount,omitempty"`
	PageSize   int    `json:"pageSize"`
	HasMore    bool   `json:"hasMore"`
}
