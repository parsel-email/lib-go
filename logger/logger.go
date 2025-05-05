// Package logger provides standardized logging functions for the auth service
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Log levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Fields type for structured logging
type Fields map[string]interface{}

// contextKey is used for context values
type contextKey int

const (
	// requestIDKey is the context key for the request ID
	requestIDKey contextKey = iota
	// traceIDKey is the context key for the trace ID
	traceIDKey
)

// Initialize sets up the global logger with proper configuration
func Initialize(level slog.Level) {
	// Set up a JSON handler for production or text handler for development
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	// Set the global logger
	slog.SetDefault(slog.New(handler))
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID gets the request ID from the context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID gets the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// contextAttrs extracts common attributes from context
func contextAttrs(ctx context.Context) []interface{} {
	var attrs []interface{}

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	// Add trace ID if available
	if traceID := GetTraceID(ctx); traceID != "" {
		attrs = append(attrs, "trace_id", traceID)
	}

	return attrs
}

// Debug logs a debug message with context
func Debug(ctx context.Context, msg string, args ...interface{}) {
	attrs := contextAttrs(ctx)
	attrs = append(attrs, args...)
	slog.Debug(msg, attrs...)
}

// Info logs an info message with context
func Info(ctx context.Context, msg string, args ...interface{}) {
	attrs := contextAttrs(ctx)
	attrs = append(attrs, args...)
	slog.Info(msg, attrs...)
}

// Warn logs a warning message with context
func Warn(ctx context.Context, msg string, args ...interface{}) {
	attrs := contextAttrs(ctx)
	attrs = append(attrs, args...)
	slog.Warn(msg, attrs...)
}

// Error logs an error message with context
func Error(ctx context.Context, msg string, args ...interface{}) {
	attrs := contextAttrs(ctx)
	attrs = append(attrs, args...)
	slog.Error(msg, attrs...)
}

// WithFields returns a new logger with the provided fields
func WithFields(fields Fields) *slog.Logger {
	attrs := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return slog.With(attrs...)
}
