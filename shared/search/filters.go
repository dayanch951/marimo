package search

import (
	"fmt"
	"strings"
	"time"
)

// FilterOperator represents comparison operators
type FilterOperator string

const (
	OpEqual        FilterOperator = "eq"
	OpNotEqual     FilterOperator = "ne"
	OpGreaterThan  FilterOperator = "gt"
	OpGreaterEqual FilterOperator = "gte"
	OpLessThan     FilterOperator = "lt"
	OpLessEqual    FilterOperator = "lte"
	OpLike         FilterOperator = "like"
	OpIn           FilterOperator = "in"
	OpNotIn        FilterOperator = "nin"
	OpBetween      FilterOperator = "between"
	OpIsNull       FilterOperator = "null"
	OpNotNull      FilterOperator = "notnull"
)

// Filter represents a single filter condition
type Filter struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
}

// FilterGroup represents a group of filters with AND/OR logic
type FilterGroup struct {
	Filters []Filter      `json:"filters"`
	Groups  []FilterGroup `json:"groups"`
	Logic   string        `json:"logic"` // "AND" or "OR"
}

// SearchRequest represents a search request with filters
type SearchRequest struct {
	Query       string      `json:"query"`        // Full-text search query
	Filters     FilterGroup `json:"filters"`      // Advanced filters
	SearchFields []string   `json:"search_fields"` // Fields to search in
}

// QueryBuilder builds SQL queries from filters
type QueryBuilder struct {
	params []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		params: make([]interface{}, 0),
	}
}

// BuildWhereClause builds WHERE clause from filter group
func (qb *QueryBuilder) BuildWhereClause(group FilterGroup) string {
	if len(group.Filters) == 0 && len(group.Groups) == 0 {
		return ""
	}

	logic := " AND "
	if group.Logic == "OR" {
		logic = " OR "
	}

	var conditions []string

	// Process filters
	for _, filter := range group.Filters {
		condition := qb.buildCondition(filter)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	// Process sub-groups
	for _, subGroup := range group.Groups {
		clause := qb.BuildWhereClause(subGroup)
		if clause != "" {
			conditions = append(conditions, fmt.Sprintf("(%s)", clause))
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, logic)
}

// buildCondition builds a single filter condition
func (qb *QueryBuilder) buildCondition(filter Filter) string {
	switch filter.Operator {
	case OpEqual:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s = $%d", filter.Field, len(qb.params))

	case OpNotEqual:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s != $%d", filter.Field, len(qb.params))

	case OpGreaterThan:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s > $%d", filter.Field, len(qb.params))

	case OpGreaterEqual:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s >= $%d", filter.Field, len(qb.params))

	case OpLessThan:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s < $%d", filter.Field, len(qb.params))

	case OpLessEqual:
		qb.params = append(qb.params, filter.Value)
		return fmt.Sprintf("%s <= $%d", filter.Field, len(qb.params))

	case OpLike:
		qb.params = append(qb.params, fmt.Sprintf("%%%v%%", filter.Value))
		return fmt.Sprintf("%s ILIKE $%d", filter.Field, len(qb.params))

	case OpIn:
		// Assuming filter.Value is a slice
		if values, ok := filter.Value.([]interface{}); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i, v := range values {
				qb.params = append(qb.params, v)
				placeholders[i] = fmt.Sprintf("$%d", len(qb.params))
			}
			return fmt.Sprintf("%s IN (%s)", filter.Field, strings.Join(placeholders, ", "))
		}
		return ""

	case OpNotIn:
		if values, ok := filter.Value.([]interface{}); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i, v := range values {
				qb.params = append(qb.params, v)
				placeholders[i] = fmt.Sprintf("$%d", len(qb.params))
			}
			return fmt.Sprintf("%s NOT IN (%s)", filter.Field, strings.Join(placeholders, ", "))
		}
		return ""

	case OpBetween:
		// Assuming filter.Value is [start, end]
		if values, ok := filter.Value.([]interface{}); ok && len(values) == 2 {
			qb.params = append(qb.params, values[0], values[1])
			return fmt.Sprintf("%s BETWEEN $%d AND $%d", filter.Field, len(qb.params)-1, len(qb.params))
		}
		return ""

	case OpIsNull:
		return fmt.Sprintf("%s IS NULL", filter.Field)

	case OpNotNull:
		return fmt.Sprintf("%s IS NOT NULL", filter.Field)

	default:
		return ""
	}
}

// BuildFullTextSearch builds full-text search clause
func (qb *QueryBuilder) BuildFullTextSearch(query string, fields []string) string {
	if query == "" || len(fields) == 0 {
		return ""
	}

	var conditions []string
	searchTerm := fmt.Sprintf("%%%s%%", query)

	for _, field := range fields {
		qb.params = append(qb.params, searchTerm)
		conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", field, len(qb.params)))
	}

	return strings.Join(conditions, " OR ")
}

// GetParams returns query parameters
func (qb *QueryBuilder) GetParams() []interface{} {
	return qb.params
}

// DateRangeFilter creates a date range filter
func DateRangeFilter(field string, from, to time.Time) Filter {
	return Filter{
		Field:    field,
		Operator: OpBetween,
		Value:    []interface{}{from, to},
	}
}

// TextSearchFilter creates a text search filter
func TextSearchFilter(field, value string) Filter {
	return Filter{
		Field:    field,
		Operator: OpLike,
		Value:    value,
	}
}

// EqualFilter creates an equality filter
func EqualFilter(field string, value interface{}) Filter {
	return Filter{
		Field:    field,
		Operator: OpEqual,
		Value:    value,
	}
}

// InFilter creates an IN filter
func InFilter(field string, values []interface{}) Filter {
	return Filter{
		Field:    field,
		Operator: OpIn,
		Value:    values,
	}
}

// BuildCompleteQuery builds a complete SQL query with search and filters
func BuildCompleteQuery(baseQuery string, searchReq SearchRequest, qb *QueryBuilder) string {
	var whereClauses []string

	// Add full-text search
	if searchReq.Query != "" && len(searchReq.SearchFields) > 0 {
		searchClause := qb.BuildFullTextSearch(searchReq.Query, searchReq.SearchFields)
		if searchClause != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)", searchClause))
		}
	}

	// Add filters
	filterClause := qb.BuildWhereClause(searchReq.Filters)
	if filterClause != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(%s)", filterClause))
	}

	// Combine query
	query := baseQuery
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return query
}
