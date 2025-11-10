package domain

const (
	DefaultLimit = 50
	MaxLimit     = 100
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// PaginatedResult holds paginated results and metadata
type PaginatedResult[T any] struct {
	Items  []T
	Total  int
	Limit  int
	Offset int
}
