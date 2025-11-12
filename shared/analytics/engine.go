package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeCount     MetricType = "count"
	MetricTypeSum       MetricType = "sum"
	MetricTypeAverage   MetricType = "average"
	MetricTypeMin       MetricType = "min"
	MetricTypeMax       MetricType = "max"
	MetricTypePercentage MetricType = "percentage"
)

// Dimension represents a dimension for grouping data
type Dimension struct {
	Name  string `json:"name"`
	Field string `json:"field"`
}

// Metric represents a calculated metric
type Metric struct {
	Name  string     `json:"name"`
	Type  MetricType `json:"type"`
	Field string     `json:"field"`
	Label string     `json:"label"`
}

// TimeRange represents a time period for analysis
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Query represents an analytics query
type Query struct {
	ID          uuid.UUID           `json:"id"`
	TenantID    uuid.UUID           `json:"tenant_id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Source      string              `json:"source"` // table or view name
	Metrics     []Metric            `json:"metrics"`
	Dimensions  []Dimension         `json:"dimensions"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	TimeRange   *TimeRange          `json:"time_range,omitempty"`
	GroupBy     []string            `json:"group_by,omitempty"`
	OrderBy     []OrderBy           `json:"order_by,omitempty"`
	Limit       int                 `json:"limit,omitempty"`
}

// OrderBy represents sorting criteria
type OrderBy struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// Result represents the result of an analytics query
type Result struct {
	Query     *Query                   `json:"query"`
	Data      []map[string]interface{} `json:"data"`
	Summary   map[string]interface{}   `json:"summary,omitempty"`
	Count     int                      `json:"count"`
	ExecTime  time.Duration            `json:"exec_time_ms"`
	CachedAt  *time.Time               `json:"cached_at,omitempty"`
}

// Engine is the analytics query engine
type Engine struct {
	db *sql.DB
}

// NewEngine creates a new analytics engine
func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// Execute runs an analytics query
func (e *Engine) Execute(ctx context.Context, query *Query) (*Result, error) {
	startTime := time.Now()

	// Build SQL query
	sqlQuery, args, err := e.buildSQL(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	// Execute query
	rows, err := e.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse results
	data, err := e.parseRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	// Calculate summary statistics
	summary := e.calculateSummary(data, query.Metrics)

	execTime := time.Since(startTime)

	return &Result{
		Query:    query,
		Data:     data,
		Summary:  summary,
		Count:    len(data),
		ExecTime: execTime,
	}, nil
}

// buildSQL builds SQL query from analytics query
func (e *Engine) buildSQL(query *Query) (string, []interface{}, error) {
	// SELECT clause
	selectClause := e.buildSelectClause(query.Metrics, query.Dimensions)

	// FROM clause
	fromClause := query.Source

	// WHERE clause
	whereClause, args := e.buildWhereClause(query.TenantID, query.Filters, query.TimeRange)

	// GROUP BY clause
	groupByClause := e.buildGroupByClause(query.Dimensions, query.GroupBy)

	// ORDER BY clause
	orderByClause := e.buildOrderByClause(query.OrderBy)

	// LIMIT clause
	limitClause := ""
	if query.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", query.Limit)
	}

	// Combine all clauses
	sqlQuery := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s %s %s %s",
		selectClause,
		fromClause,
		whereClause,
		groupByClause,
		orderByClause,
		limitClause,
	)

	return sqlQuery, args, nil
}

// buildSelectClause builds SELECT part of SQL
func (e *Engine) buildSelectClause(metrics []Metric, dimensions []Dimension) string {
	var parts []string

	// Add dimensions
	for _, dim := range dimensions {
		parts = append(parts, fmt.Sprintf("%s AS %s", dim.Field, dim.Name))
	}

	// Add metrics
	for _, metric := range metrics {
		var expr string
		switch metric.Type {
		case MetricTypeCount:
			expr = fmt.Sprintf("COUNT(%s)", metric.Field)
		case MetricTypeSum:
			expr = fmt.Sprintf("SUM(%s)", metric.Field)
		case MetricTypeAverage:
			expr = fmt.Sprintf("AVG(%s)", metric.Field)
		case MetricTypeMin:
			expr = fmt.Sprintf("MIN(%s)", metric.Field)
		case MetricTypeMax:
			expr = fmt.Sprintf("MAX(%s)", metric.Field)
		default:
			expr = metric.Field
		}
		parts = append(parts, fmt.Sprintf("%s AS %s", expr, metric.Name))
	}

	return joinStrings(parts, ", ")
}

// buildWhereClause builds WHERE part of SQL
func (e *Engine) buildWhereClause(tenantID uuid.UUID, filters map[string]interface{}, timeRange *TimeRange) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add tenant filter
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID)
	argIndex++

	// Add custom filters
	for field, value := range filters {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Add time range filter
	if timeRange != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, timeRange.Start)
		argIndex++

		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, timeRange.End)
		argIndex++
	}

	return joinStrings(conditions, " AND "), args
}

// buildGroupByClause builds GROUP BY part of SQL
func (e *Engine) buildGroupByClause(dimensions []Dimension, groupBy []string) string {
	if len(dimensions) == 0 && len(groupBy) == 0 {
		return ""
	}

	var fields []string
	for _, dim := range dimensions {
		fields = append(fields, dim.Field)
	}
	fields = append(fields, groupBy...)

	if len(fields) > 0 {
		return "GROUP BY " + joinStrings(fields, ", ")
	}
	return ""
}

// buildOrderByClause builds ORDER BY part of SQL
func (e *Engine) buildOrderByClause(orderBy []OrderBy) string {
	if len(orderBy) == 0 {
		return ""
	}

	var parts []string
	for _, ob := range orderBy {
		direction := "ASC"
		if ob.Desc {
			direction = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s", ob.Field, direction))
	}

	return "ORDER BY " + joinStrings(parts, ", ")
}

// parseRows converts SQL rows to map slice
func (e *Engine) parseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	return results, rows.Err()
}

// calculateSummary calculates summary statistics
func (e *Engine) calculateSummary(data []map[string]interface{}, metrics []Metric) map[string]interface{} {
	summary := make(map[string]interface{})

	for _, metric := range metrics {
		var sum float64
		var count int
		var min, max float64
		first := true

		for _, row := range data {
			if val, ok := row[metric.Name]; ok {
				var floatVal float64
				switch v := val.(type) {
				case float64:
					floatVal = v
				case int:
					floatVal = float64(v)
				case int64:
					floatVal = float64(v)
				default:
					continue
				}

				sum += floatVal
				count++

				if first {
					min = floatVal
					max = floatVal
					first = false
				} else {
					if floatVal < min {
						min = floatVal
					}
					if floatVal > max {
						max = floatVal
					}
				}
			}
		}

		if count > 0 {
			summary[metric.Name+"_total"] = sum
			summary[metric.Name+"_avg"] = sum / float64(count)
			summary[metric.Name+"_min"] = min
			summary[metric.Name+"_max"] = max
			summary[metric.Name+"_count"] = count
		}
	}

	return summary
}

// Helper function to join strings
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// SaveQuery saves a query for reuse
func (e *Engine) SaveQuery(ctx context.Context, query *Query) error {
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return err
	}

	sqlQuery := `
		INSERT INTO analytics_queries (id, tenant_id, name, description, query_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET query_json = $5, updated_at = $7
	`

	now := time.Now()
	_, err = e.db.ExecContext(ctx, sqlQuery,
		query.ID, query.TenantID, query.Name, query.Description,
		queryJSON, now, now,
	)

	return err
}

// LoadQuery loads a saved query
func (e *Engine) LoadQuery(ctx context.Context, queryID uuid.UUID) (*Query, error) {
	sqlQuery := `
		SELECT query_json
		FROM analytics_queries
		WHERE id = $1
	`

	var queryJSON []byte
	err := e.db.QueryRowContext(ctx, sqlQuery, queryID).Scan(&queryJSON)
	if err != nil {
		return nil, err
	}

	var query Query
	if err := json.Unmarshal(queryJSON, &query); err != nil {
		return nil, err
	}

	return &query, nil
}
