package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReportType defines the type of report
type ReportType string

const (
	ReportTypeUserActivity    ReportType = "user_activity"
	ReportTypeRevenue         ReportType = "revenue"
	ReportTypeUsage           ReportType = "usage"
	ReportTypePerformance     ReportType = "performance"
	ReportTypeCustom          ReportType = "custom"
)

// ReportSchedule defines when a report should be generated
type ReportSchedule string

const (
	ScheduleDaily   ReportSchedule = "daily"
	ScheduleWeekly  ReportSchedule = "weekly"
	ScheduleMonthly ReportSchedule = "monthly"
	ScheduleCustom  ReportSchedule = "custom"
)

// Report represents a configured report
type Report struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Name        string         `json:"name"`
	Type        ReportType     `json:"type"`
	Query       *Query         `json:"query"`
	Schedule    ReportSchedule `json:"schedule"`
	Recipients  []string       `json:"recipients"` // Email addresses
	Format      string         `json:"format"` // pdf, csv, excel
	Enabled     bool           `json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastRunAt   *time.Time     `json:"last_run_at,omitempty"`
	NextRunAt   *time.Time     `json:"next_run_at,omitempty"`
}

// ReportBuilder helps build predefined reports
type ReportBuilder struct {
	engine *Engine
}

// NewReportBuilder creates a new report builder
func NewReportBuilder(engine *Engine) *ReportBuilder {
	return &ReportBuilder{engine: engine}
}

// BuildUserActivityReport creates a user activity report
func (rb *ReportBuilder) BuildUserActivityReport(ctx context.Context, tenantID uuid.UUID, timeRange TimeRange) (*Result, error) {
	query := &Query{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "User Activity Report",
		Source:   "user_activities",
		Metrics: []Metric{
			{Name: "total_users", Type: MetricTypeCount, Field: "DISTINCT user_id", Label: "Total Active Users"},
			{Name: "total_actions", Type: MetricTypeCount, Field: "*", Label: "Total Actions"},
			{Name: "avg_session_duration", Type: MetricTypeAverage, Field: "session_duration", Label: "Avg Session Duration"},
		},
		Dimensions: []Dimension{
			{Name: "date", Field: "DATE(created_at)"},
			{Name: "action_type", Field: "action_type"},
		},
		TimeRange: &timeRange,
		GroupBy:   []string{"DATE(created_at)", "action_type"},
		OrderBy: []OrderBy{
			{Field: "DATE(created_at)", Desc: true},
		},
	}

	return rb.engine.Execute(ctx, query)
}

// BuildRevenueReport creates a revenue report
func (rb *ReportBuilder) BuildRevenueReport(ctx context.Context, tenantID uuid.UUID, timeRange TimeRange) (*Result, error) {
	query := &Query{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "Revenue Report",
		Source:   "transactions",
		Metrics: []Metric{
			{Name: "total_revenue", Type: MetricTypeSum, Field: "amount", Label: "Total Revenue"},
			{Name: "transaction_count", Type: MetricTypeCount, Field: "*", Label: "Transaction Count"},
			{Name: "avg_transaction", Type: MetricTypeAverage, Field: "amount", Label: "Average Transaction"},
			{Name: "max_transaction", Type: MetricTypeMax, Field: "amount", Label: "Largest Transaction"},
		},
		Dimensions: []Dimension{
			{Name: "date", Field: "DATE(created_at)"},
			{Name: "payment_method", Field: "payment_method"},
		},
		TimeRange: &timeRange,
		GroupBy:   []string{"DATE(created_at)", "payment_method"},
		OrderBy: []OrderBy{
			{Field: "DATE(created_at)", Desc: true},
		},
	}

	return rb.engine.Execute(ctx, query)
}

// BuildUsageReport creates a system usage report
func (rb *ReportBuilder) BuildUsageReport(ctx context.Context, tenantID uuid.UUID, timeRange TimeRange) (*Result, error) {
	query := &Query{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "Usage Report",
		Source:   "usage_metrics",
		Metrics: []Metric{
			{Name: "api_calls", Type: MetricTypeSum, Field: "api_calls", Label: "API Calls"},
			{Name: "storage_used", Type: MetricTypeSum, Field: "storage_bytes", Label: "Storage Used (bytes)"},
			{Name: "active_users", Type: MetricTypeCount, Field: "DISTINCT user_id", Label: "Active Users"},
		},
		Dimensions: []Dimension{
			{Name: "date", Field: "DATE(created_at)"},
		},
		TimeRange: &timeRange,
		GroupBy:   []string{"DATE(created_at)"},
		OrderBy: []OrderBy{
			{Field: "DATE(created_at)", Desc: true},
		},
	}

	return rb.engine.Execute(ctx, query)
}

// BuildPerformanceReport creates a performance report
func (rb *ReportBuilder) BuildPerformanceReport(ctx context.Context, tenantID uuid.UUID, timeRange TimeRange) (*Result, error) {
	query := &Query{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "Performance Report",
		Source:   "performance_metrics",
		Metrics: []Metric{
			{Name: "avg_response_time", Type: MetricTypeAverage, Field: "response_time_ms", Label: "Avg Response Time (ms)"},
			{Name: "max_response_time", Type: MetricTypeMax, Field: "response_time_ms", Label: "Max Response Time (ms)"},
			{Name: "error_rate", Type: MetricTypePercentage, Field: "errors", Label: "Error Rate (%)"},
			{Name: "request_count", Type: MetricTypeCount, Field: "*", Label: "Request Count"},
		},
		Dimensions: []Dimension{
			{Name: "endpoint", Field: "endpoint"},
			{Name: "method", Field: "method"},
		},
		TimeRange: &timeRange,
		GroupBy:   []string{"endpoint", "method"},
		OrderBy: []OrderBy{
			{Field: "avg_response_time", Desc: true},
		},
		Limit: 100,
	}

	return rb.engine.Execute(ctx, query)
}

// Dashboard represents a collection of widgets
type Dashboard struct {
	ID        uuid.UUID        `json:"id"`
	TenantID  uuid.UUID        `json:"tenant_id"`
	Name      string           `json:"name"`
	Widgets   []Widget         `json:"widgets"`
	Layout    DashboardLayout  `json:"layout"`
	IsPublic  bool             `json:"is_public"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Widget represents a single visualization on a dashboard
type Widget struct {
	ID           string       `json:"id"`
	Type         WidgetType   `json:"type"`
	Title        string       `json:"title"`
	Query        *Query       `json:"query"`
	Visualization string      `json:"visualization"` // line, bar, pie, table, metric
	Settings     WidgetSettings `json:"settings"`
	Position     Position     `json:"position"`
}

// WidgetType defines the type of widget
type WidgetType string

const (
	WidgetTypeMetric WidgetType = "metric"
	WidgetTypeChart  WidgetType = "chart"
	WidgetTypeTable  WidgetType = "table"
	WidgetTypeCustom WidgetType = "custom"
)

// WidgetSettings contains widget-specific configuration
type WidgetSettings struct {
	RefreshInterval int                    `json:"refresh_interval"` // seconds
	Colors          []string               `json:"colors,omitempty"`
	ShowLegend      bool                   `json:"show_legend"`
	ShowGrid        bool                   `json:"show_grid"`
	Custom          map[string]interface{} `json:"custom,omitempty"`
}

// Position defines widget position on dashboard
type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DashboardLayout defines dashboard grid layout
type DashboardLayout struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// DashboardService manages dashboards
type DashboardService struct {
	engine *Engine
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(engine *Engine) *DashboardService {
	return &DashboardService{engine: engine}
}

// CreateDefaultDashboard creates a default dashboard for new tenants
func (ds *DashboardService) CreateDefaultDashboard(tenantID uuid.UUID) *Dashboard {
	return &Dashboard{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "Overview",
		Layout: DashboardLayout{
			Columns: 12,
			Rows:    6,
		},
		Widgets: []Widget{
			{
				ID:    "total-users",
				Type:  WidgetTypeMetric,
				Title: "Total Users",
				Query: &Query{
					Source: "users",
					Metrics: []Metric{
						{Name: "count", Type: MetricTypeCount, Field: "*"},
					},
				},
				Visualization: "metric",
				Position: Position{X: 0, Y: 0, Width: 3, Height: 2},
			},
			{
				ID:    "revenue-chart",
				Type:  WidgetTypeChart,
				Title: "Revenue Trend",
				Query: &Query{
					Source: "transactions",
					Metrics: []Metric{
						{Name: "revenue", Type: MetricTypeSum, Field: "amount"},
					},
					Dimensions: []Dimension{
						{Name: "date", Field: "DATE(created_at)"},
					},
					TimeRange: &TimeRange{
						Start: time.Now().AddDate(0, -1, 0),
						End:   time.Now(),
					},
				},
				Visualization: "line",
				Settings: WidgetSettings{
					ShowLegend: true,
					ShowGrid:   true,
				},
				Position: Position{X: 3, Y: 0, Width: 6, Height: 4},
			},
			{
				ID:    "recent-activities",
				Type:  WidgetTypeTable,
				Title: "Recent Activities",
				Query: &Query{
					Source: "user_activities",
					Dimensions: []Dimension{
						{Name: "user", Field: "user_name"},
						{Name: "action", Field: "action_type"},
						{Name: "timestamp", Field: "created_at"},
					},
					OrderBy: []OrderBy{
						{Field: "created_at", Desc: true},
					},
					Limit: 10,
				},
				Visualization: "table",
				Position: Position{X: 9, Y: 0, Width: 3, Height: 4},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// RenderDashboard executes all queries in a dashboard
func (ds *DashboardService) RenderDashboard(ctx context.Context, dashboard *Dashboard) (map[string]*Result, error) {
	results := make(map[string]*Result)

	for _, widget := range dashboard.Widgets {
		if widget.Query == nil {
			continue
		}

		// Set tenant ID
		widget.Query.TenantID = dashboard.TenantID

		result, err := ds.engine.Execute(ctx, widget.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query for widget %s: %w", widget.ID, err)
		}

		results[widget.ID] = result
	}

	return results, nil
}
