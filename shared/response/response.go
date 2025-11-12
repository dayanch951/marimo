package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"marimo/shared/errors"
)

// Response represents the standard API response structure
type Response struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     *ErrorDetails          `json:"error,omitempty"`
	Meta      *Meta                  `json:"meta,omitempty"`
	Timestamp string                 `json:"timestamp"`
	TraceID   string                 `json:"trace_id"`
}

// ErrorDetails represents error information
type ErrorDetails struct {
	Code    errors.ErrorCode       `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Meta represents metadata for responses (pagination, etc.)
type Meta struct {
	Page       int         `json:"page,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Total      int64       `json:"total,omitempty"`
	TotalPages int         `json:"total_pages,omitempty"`
	Extra      interface{} `json:"extra,omitempty"`
}

// PaginationMeta creates pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// getTraceID gets or creates trace ID from context
func getTraceID(c *gin.Context) string {
	// Try to get existing trace ID from context
	if traceID, exists := c.Get("trace_id"); exists {
		return traceID.(string)
	}

	// Try to get from request header
	traceID := c.GetHeader("X-Trace-ID")
	if traceID == "" {
		traceID = uuid.New().String()
	}

	c.Set("trace_id", traceID)
	return traceID
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, err *errors.AppError) {
	c.JSON(err.StatusCode, Response{
		Success: false,
		Error: &ErrorDetails{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	})
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, pagination *PaginationMeta) {
	meta := &Meta{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      pagination.Total,
		TotalPages: pagination.TotalPages,
	}

	SuccessWithMeta(c, data, meta)
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) *PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// List represents a list response with pagination
type List struct {
	Items      interface{}      `json:"items"`
	Pagination *PaginationMeta `json:"pagination"`
}

// NewList creates a new list response
func NewList(items interface{}, pagination *PaginationMeta) *List {
	return &List{
		Items:      items,
		Pagination: pagination,
	}
}

// PaginatedList sends a list response with pagination
func PaginatedList(c *gin.Context, items interface{}, pagination *PaginationMeta) {
	Success(c, NewList(items, pagination))
}

// Message sends a simple message response
func Message(c *gin.Context, message string) {
	Success(c, gin.H{"message": message})
}

// MessageWithStatus sends a message with custom status code
func MessageWithStatus(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      gin.H{"message": message},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	})
}

// Custom response builders
type ResponseBuilder struct {
	success   bool
	data      interface{}
	err       *errors.AppError
	meta      *Meta
	statusCode int
}

// NewResponse creates a new response builder
func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		success:    true,
		statusCode: http.StatusOK,
	}
}

// WithData sets the response data
func (rb *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	rb.data = data
	return rb
}

// WithError sets the error
func (rb *ResponseBuilder) WithError(err *errors.AppError) *ResponseBuilder {
	rb.success = false
	rb.err = err
	rb.statusCode = err.StatusCode
	return rb
}

// WithMeta sets metadata
func (rb *ResponseBuilder) WithMeta(meta *Meta) *ResponseBuilder {
	rb.meta = meta
	return rb
}

// WithStatus sets custom status code
func (rb *ResponseBuilder) WithStatus(code int) *ResponseBuilder {
	rb.statusCode = code
	return rb
}

// Send sends the response
func (rb *ResponseBuilder) Send(c *gin.Context) {
	response := Response{
		Success:   rb.success,
		Data:      rb.data,
		Meta:      rb.meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TraceID:   getTraceID(c),
	}

	if rb.err != nil {
		response.Error = &ErrorDetails{
			Code:    rb.err.Code,
			Message: rb.err.Message,
			Details: rb.err.Details,
		}
	}

	c.JSON(rb.statusCode, response)
}
