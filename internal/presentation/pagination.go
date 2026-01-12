package presentation

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

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

// NormalizePagination normalizes pagination parameters
func NormalizePagination(page, limit int) (offset, normalizedLimit int) {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset = (page - 1) * limit
	return offset, limit
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(total, limit int) int {
	if total == 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}
