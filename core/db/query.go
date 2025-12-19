package db

import (
	"fmt"
	"strings"
)

// QueryBuilder builds SQL queries with fluent API
type QueryBuilder struct {
	tableName   string
	columns     []string
	conditions  []string
	args        []any
	orderBy     []string
	groupBy     []string
	having      []string
	havingArgs  []any
	limit       int
	offset      int
	withDeleted bool
	countOnly   bool
	joins       []string
}

// NewQueryBuilder creates a new QueryBuilder for a table
func NewQueryBuilder(tableName string) *QueryBuilder {
	return &QueryBuilder{
		tableName: tableName,
	}
}

// QueryOption is a function that modifies a QueryBuilder
type QueryOption func(*QueryBuilder)

// Where adds a WHERE condition
func Where(condition string, args ...any) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, condition)
		qb.args = append(qb.args, args...)
	}
}

// WhereEq adds a WHERE column = value condition
func WhereEq(column string, value any) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s = ?", column))
		qb.args = append(qb.args, value)
	}
}

// WhereIn adds a WHERE column IN (...) condition
func WhereIn(column string, values ...any) QueryOption {
	return func(qb *QueryBuilder) {
		if len(values) == 0 {
			return
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
		qb.args = append(qb.args, values...)
	}
}

// WhereNotNull adds a WHERE column IS NOT NULL condition
func WhereNotNull(column string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s IS NOT NULL", column))
	}
}

// WhereNull adds a WHERE column IS NULL condition
func WhereNull(column string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s IS NULL", column))
	}
}

// WhereLike adds a WHERE column LIKE pattern condition
func WhereLike(column string, pattern string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s LIKE ?", column))
		qb.args = append(qb.args, pattern)
	}
}

// WhereBetween adds a WHERE column BETWEEN min AND max condition
func WhereBetween(column string, min, max any) QueryOption {
	return func(qb *QueryBuilder) {
		qb.conditions = append(qb.conditions, fmt.Sprintf("%s BETWEEN ? AND ?", column))
		qb.args = append(qb.args, min, max)
	}
}

// OrderBy adds an ORDER BY clause
func OrderBy(column, direction string) QueryOption {
	return func(qb *QueryBuilder) {
		dir := strings.ToUpper(strings.TrimSpace(direction))
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, dir))
	}
}

// OrderByAsc adds an ORDER BY column ASC clause
func OrderByAsc(column string) QueryOption {
	return OrderBy(column, "ASC")
}

// OrderByDesc adds an ORDER BY column DESC clause
func OrderByDesc(column string) QueryOption {
	return OrderBy(column, "DESC")
}

// GroupBy adds a GROUP BY clause
func GroupBy(columns ...string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.groupBy = append(qb.groupBy, columns...)
	}
}

// Having adds a HAVING condition
func Having(condition string, args ...any) QueryOption {
	return func(qb *QueryBuilder) {
		qb.having = append(qb.having, condition)
		qb.havingArgs = append(qb.havingArgs, args...)
	}
}

// Limit sets the LIMIT clause
func Limit(n int) QueryOption {
	return func(qb *QueryBuilder) {
		if n > 0 {
			qb.limit = n
		}
	}
}

// Offset sets the OFFSET clause
func Offset(n int) QueryOption {
	return func(qb *QueryBuilder) {
		if n >= 0 {
			qb.offset = n
		}
	}
}

// Paginate sets both LIMIT and OFFSET for pagination
func Paginate(page, perPage int) QueryOption {
	return func(qb *QueryBuilder) {
		if page < 1 {
			page = 1
		}
		if perPage < 1 {
			perPage = 10
		}
		qb.limit = perPage
		qb.offset = (page - 1) * perPage
	}
}

// WithDeleted includes soft-deleted records
func WithDeleted() QueryOption {
	return func(qb *QueryBuilder) {
		qb.withDeleted = true
	}
}

// OnlyDeleted returns only soft-deleted records
func OnlyDeleted() QueryOption {
	return func(qb *QueryBuilder) {
		qb.withDeleted = true
		qb.conditions = append(qb.conditions, "deleted_at IS NOT NULL")
	}
}

// Select specifies columns to select
func Select(columns ...string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.columns = append(qb.columns, columns...)
	}
}

// Join adds a JOIN clause
func Join(join string) QueryOption {
	return func(qb *QueryBuilder) {
		qb.joins = append(qb.joins, join)
	}
}

// Apply applies QueryOptions to the builder
func (qb *QueryBuilder) Apply(opts ...QueryOption) *QueryBuilder {
	for _, opt := range opts {
		opt(qb)
	}
	return qb
}

// Build generates the SQL query and arguments
func (qb *QueryBuilder) Build() (string, []any) {
	var query strings.Builder
	var allArgs []any

	// SELECT clause
	if qb.countOnly {
		query.WriteString("SELECT COUNT(*) FROM ")
	} else {
		query.WriteString("SELECT ")
		if len(qb.columns) > 0 {
			query.WriteString(strings.Join(qb.columns, ", "))
		} else {
			query.WriteString("*")
		}
		query.WriteString(" FROM ")
	}

	query.WriteString(qb.tableName)

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(join)
	}

	// WHERE clause - add soft delete filter first
	conditions := make([]string, 0, len(qb.conditions)+1)
	if !qb.withDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}
	conditions = append(conditions, qb.conditions...)

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}
	allArgs = append(allArgs, qb.args...)

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		query.WriteString(strings.Join(qb.having, " AND "))
		allArgs = append(allArgs, qb.havingArgs...)
	}

	// ORDER BY, LIMIT, OFFSET only for non-count queries
	if !qb.countOnly {
		if len(qb.orderBy) > 0 {
			query.WriteString(" ORDER BY ")
			query.WriteString(strings.Join(qb.orderBy, ", "))
		}

		if qb.limit > 0 {
			query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
		} else if qb.offset > 0 {
			// SQLite requires LIMIT when using OFFSET
			query.WriteString(" LIMIT -1")
		}

		if qb.offset > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
		}
	}

	return query.String(), allArgs
}

// BuildCount generates a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []any) {
	qb.countOnly = true
	return qb.Build()
}

// Clone creates a copy of the QueryBuilder
func (qb *QueryBuilder) Clone() *QueryBuilder {
	clone := &QueryBuilder{
		tableName:   qb.tableName,
		columns:     append([]string{}, qb.columns...),
		conditions:  append([]string{}, qb.conditions...),
		args:        append([]any{}, qb.args...),
		orderBy:     append([]string{}, qb.orderBy...),
		groupBy:     append([]string{}, qb.groupBy...),
		having:      append([]string{}, qb.having...),
		havingArgs:  append([]any{}, qb.havingArgs...),
		limit:       qb.limit,
		offset:      qb.offset,
		withDeleted: qb.withDeleted,
		countOnly:   qb.countOnly,
		joins:       append([]string{}, qb.joins...),
	}
	return clone
}

// Reset clears all query options
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.columns = nil
	qb.conditions = nil
	qb.args = nil
	qb.orderBy = nil
	qb.groupBy = nil
	qb.having = nil
	qb.havingArgs = nil
	qb.limit = 0
	qb.offset = 0
	qb.withDeleted = false
	qb.countOnly = false
	qb.joins = nil
	return qb
}

// TableName returns the table name
func (qb *QueryBuilder) TableName() string {
	return qb.tableName
}
