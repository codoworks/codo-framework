// Package pagination provides pagination parameter extraction and helpers
// for building paginated API responses.
package pagination

// Type represents the pagination strategy
type Type string

const (
	// TypeOffset uses page/offset-based pagination
	TypeOffset Type = "offset"
	// TypeCursor uses cursor-based pagination
	TypeCursor Type = "cursor"
)

// Direction represents cursor navigation direction
type Direction string

const (
	// DirectionNext navigates forward through results
	DirectionNext Direction = "next"
	// DirectionPrev navigates backward through results
	DirectionPrev Direction = "prev"
)

// Params holds validated pagination parameters extracted from the request.
// These parameters are set by the pagination middleware and can be retrieved
// using Get() or GetOrDefault().
type Params struct {
	// Type indicates the pagination strategy (offset or cursor)
	Type Type

	// PerPage is the number of items per page (validated, within max)
	PerPage int

	// MaxPerPage is the maximum allowed per_page value (from config)
	MaxPerPage int

	// Page is the 1-indexed page number (offset-based only)
	Page int

	// Offset is the calculated offset: (Page - 1) * PerPage (offset-based only)
	Offset int

	// Cursor is the opaque cursor string (cursor-based only)
	Cursor string

	// Direction is the navigation direction (cursor-based only)
	Direction Direction

	// Raw values for debugging/logging
	RawPage    string
	RawPerPage string
	RawCursor  string
}

// IsOffset returns true if using offset-based pagination
func (p *Params) IsOffset() bool {
	return p.Type == TypeOffset
}

// IsCursor returns true if using cursor-based pagination
func (p *Params) IsCursor() bool {
	return p.Type == TypeCursor
}

// HasCursor returns true if a cursor was provided in the request
func (p *Params) HasCursor() bool {
	return p.Cursor != ""
}

// Limit returns the per-page value (alias for PerPage for clarity)
func (p *Params) Limit() int {
	return p.PerPage
}

// WithMaxPerPage returns a copy with a different max per page.
// Useful for endpoints that need a different limit than the global config.
// If the current PerPage exceeds the new max, it will be capped.
func (p *Params) WithMaxPerPage(max int) *Params {
	if p == nil {
		return nil
	}

	clone := *p
	clone.MaxPerPage = max

	// Re-validate per_page against new max
	if clone.PerPage > max {
		clone.PerPage = max
		// Recalculate offset for offset-based pagination
		if clone.Type == TypeOffset {
			clone.Offset = (clone.Page - 1) * clone.PerPage
		}
	}

	return &clone
}

// WithPerPage returns a copy with a different per_page value.
// The value is capped at MaxPerPage and must be at least 1.
func (p *Params) WithPerPage(perPage int) *Params {
	if p == nil {
		return nil
	}

	clone := *p
	if perPage > clone.MaxPerPage {
		perPage = clone.MaxPerPage
	}
	if perPage < 1 {
		perPage = 1
	}
	clone.PerPage = perPage

	// Recalculate offset for offset-based pagination
	if clone.Type == TypeOffset {
		clone.Offset = (clone.Page - 1) * perPage
	}

	return &clone
}

// WithPage returns a copy with a different page number.
// The page must be at least 1.
func (p *Params) WithPage(page int) *Params {
	if p == nil {
		return nil
	}

	clone := *p
	if page < 1 {
		page = 1
	}
	clone.Page = page
	clone.Offset = (page - 1) * clone.PerPage

	return &clone
}
