package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"marimo/shared/errors"
	"marimo/shared/response"
)

// ErrorHandler middleware handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return 500 Internal Server Error
				response.Error(c, errors.Internal("An unexpected error occurred"))
				c.Abort()
			}
		}()

		c.Next()

		// Handle errors from handlers
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Check if it's an AppError
			if appErr, ok := errors.GetAppError(err); ok {
				response.Error(c, appErr)
				return
			}

			// Unknown error - return as internal error
			log.Printf("Unhandled error: %v", err)
			response.Error(c, errors.Internal(err.Error()))
		}
	}
}

// RecoveryMiddleware catches panics and returns proper error response
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				// Return internal server error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  errors.ErrInternal,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, errors.NotFound(fmt.Sprintf("Route not found: %s %s", c.Request.Method, c.Request.URL.Path)))
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, errors.New(errors.ErrBadRequest, fmt.Sprintf("Method %s not allowed for %s", c.Request.Method, c.Request.URL.Path)))
	}
}
