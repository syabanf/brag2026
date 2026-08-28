package domain

// Default and ceiling for a page. The ceiling matters: without it a caller can
// ask for every row and turn a list endpoint into a denial of service.
const (
	DefaultPageSize = 25
	MaxPageSize     = 200
)

// Page is a window over a larger result set.
type Page struct {
	Limit  int
	Offset int
}

// Normalise clamps a page into the allowed range, so a handler can pass query
// parameters straight through without validating each one.
func (p Page) Normalise() Page {
	if p.Limit <= 0 {
		p.Limit = DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		p.Limit = MaxPageSize
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

// Paged carries the total alongside the rows so a UI can render "1–25 of 180"
// and know whether another page exists.
type Paged[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func NewPaged[T any](items []T, total int, page Page) Paged[T] {
	if items == nil {
		items = []T{}
	}
	return Paged[T]{Items: items, Total: total, Limit: page.Limit, Offset: page.Offset}
}

// HasMore reports whether another page follows this one.
func (p Paged[T]) HasMore() bool {
	return p.Offset+len(p.Items) < p.Total
}
