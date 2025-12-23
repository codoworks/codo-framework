package pagination

import (
	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/db"
)

// QueryOptions returns db.QueryOption slice from context pagination params.
// Returns nil if no pagination is in context.
// This is a convenience function that combines Get() and Params.QueryOptions().
func QueryOptions(c echo.Context) []db.QueryOption {
	params := Get(c)
	if params == nil {
		return nil
	}
	return params.QueryOptions()
}

// QueryOptions converts pagination params to db.QueryOption slice.
// For offset-based pagination, returns Limit and Offset options.
// For cursor-based pagination, returns Limit+1 (to detect HasMore).
func (p *Params) QueryOptions() []db.QueryOption {
	if p == nil {
		return nil
	}

	// For offset-based, use Limit and Offset
	if p.IsOffset() {
		return []db.QueryOption{
			db.Limit(p.PerPage),
			db.Offset(p.Offset),
		}
	}

	// For cursor-based, fetch one extra to determine HasMore
	// Caller handles cursor decoding and WHERE clause
	return []db.QueryOption{
		db.Limit(p.PerPage + 1),
	}
}

// QueryOptionsWithOrder returns QueryOptions plus an order option.
// Useful for ensuring consistent ordering with pagination.
func (p *Params) QueryOptionsWithOrder(column, direction string) []db.QueryOption {
	if p == nil {
		return nil
	}

	opts := p.QueryOptions()
	opts = append(opts, db.OrderBy(column, direction))
	return opts
}

// QueryOptionsWithOrderAsc returns QueryOptions plus an ascending order option.
func (p *Params) QueryOptionsWithOrderAsc(column string) []db.QueryOption {
	return p.QueryOptionsWithOrder(column, "ASC")
}

// QueryOptionsWithOrderDesc returns QueryOptions plus a descending order option.
func (p *Params) QueryOptionsWithOrderDesc(column string) []db.QueryOption {
	return p.QueryOptionsWithOrder(column, "DESC")
}
