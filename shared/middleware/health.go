package middleware

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status      string                 `json:"status"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]CheckResult `json:"checks"`
	System      SystemInfo             `json:"system"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// SystemInfo contains system resource information
type SystemInfo struct {
	GoVersion      string  `json:"go_version"`
	NumGoroutines  int     `json:"num_goroutines"`
	MemAllocMB     float64 `json:"mem_alloc_mb"`
	MemTotalMB     float64 `json:"mem_total_mb"`
	NumCPU         int     `json:"num_cpu"`
}

var startTime = time.Now()

// HealthCheckHandler creates a detailed health check endpoint
func HealthCheckHandler(serviceName, version string, checks map[string]func() CheckResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := HealthStatus{
			Status:    "healthy",
			Service:   serviceName,
			Version:   version,
			Timestamp: time.Now(),
			Uptime:    time.Since(startTime).String(),
			Checks:    make(map[string]CheckResult),
			System:    getSystemInfo(),
		}

		// Run all health checks
		allHealthy := true
		for name, check := range checks {
			result := check()
			status.Checks[name] = result
			if result.Status != "healthy" {
				allHealthy = false
			}
		}

		if !allHealthy {
			status.Status = "degraded"
		}

		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if status.Status == "degraded" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(status)
	}
}

// getSystemInfo returns current system resource information
func getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:      runtime.Version(),
		NumGoroutines:  runtime.NumGoroutine(),
		MemAllocMB:     float64(m.Alloc) / 1024 / 1024,
		MemTotalMB:     float64(m.TotalAlloc) / 1024 / 1024,
		NumCPU:         runtime.NumCPU(),
	}
}

// DatabaseHealthCheck creates a health check for database connectivity
func DatabaseHealthCheck(pingFunc func() error) func() CheckResult {
	return func() CheckResult {
		start := time.Now()
		err := pingFunc()
		latency := time.Since(start)

		if err != nil {
			return CheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
				Latency: latency.String(),
			}
		}

		return CheckResult{
			Status:  "healthy",
			Latency: latency.String(),
		}
	}
}

// DependencyHealthCheck creates a health check for external dependencies
func DependencyHealthCheck(name, url string) func() CheckResult {
	return func() CheckResult {
		start := time.Now()
		client := &http.Client{Timeout: 5 * time.Second}

		resp, err := client.Get(url)
		latency := time.Since(start)

		if err != nil {
			return CheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
				Latency: latency.String(),
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return CheckResult{
				Status:  "healthy",
				Latency: latency.String(),
			}
		}

		return CheckResult{
			Status:  "unhealthy",
			Message: "HTTP " + resp.Status,
			Latency: latency.String(),
		}
	}
}
