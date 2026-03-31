package core

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
