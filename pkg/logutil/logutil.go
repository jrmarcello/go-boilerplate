package logutil

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// Context key type for safe context value storage.
type contextKey string

const logContextKey contextKey = "log_context"

// Step constants for identifying the layer in logs.
const (
	StepHandler    = "handler"
	StepUseCase    = "usecase"
	StepRepository = "repository"
	StepCache      = "cache"
	StepMiddleware = "middleware"
)

// LogContext holds structured logging context across layers.
type LogContext struct {
	RequestID     string
	TraceID       string
	UserID        string
	CallerService string
	Step          string
	Resource      string
	Action        string
	Extra         map[string]any
}

// Inject stores a LogContext in the context.
func Inject(ctx context.Context, lc LogContext) context.Context {
	return context.WithValue(ctx, logContextKey, lc)
}

// Extract retrieves a LogContext from the context.
func Extract(ctx context.Context) (LogContext, bool) {
	lc, ok := ctx.Value(logContextKey).(LogContext)
	return lc, ok
}

// WithStep returns a new LogContext copy with the step set.
func (lc LogContext) WithStep(step string) LogContext {
	lc.Step = step
	return lc
}

// WithResource returns a new LogContext copy with the resource set.
func (lc LogContext) WithResource(resource string) LogContext {
	lc.Resource = resource
	return lc
}

// WithAction returns a new LogContext copy with the action set.
func (lc LogContext) WithAction(action string) LogContext {
	lc.Action = action
	return lc
}

// ToSlogAttrs converts the LogContext to slog attributes for structured logging.
func (lc LogContext) ToSlogAttrs() []slog.Attr {
	attrs := []slog.Attr{}

	if lc.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", lc.RequestID))
	}
	if lc.TraceID != "" {
		attrs = append(attrs, slog.String("trace_id", lc.TraceID))
	}
	if lc.UserID != "" {
		attrs = append(attrs, slog.String("user_id", lc.UserID))
	}
	if lc.CallerService != "" {
		attrs = append(attrs, slog.String("caller_service", lc.CallerService))
	}
	if lc.Step != "" {
		attrs = append(attrs, slog.String("step", lc.Step))
	}
	if lc.Resource != "" {
		attrs = append(attrs, slog.String("resource", lc.Resource))
	}
	if lc.Action != "" {
		attrs = append(attrs, slog.String("action", lc.Action))
	}

	for k, v := range lc.Extra {
		attrs = append(attrs, slog.Any(k, v))
	}

	return attrs
}

// ErrorLogFields returns slog attributes for error logging.
// Includes stack trace for 5xx errors (non-domain errors).
func ErrorLogFields(err error, isDomainError bool) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("error", err.Error()),
		slog.String("error_type", fmt.Sprintf("%T", err)),
	}

	if !isDomainError {
		attrs = append(attrs, slog.String("stack_trace", getStackTrace()))
	}

	return attrs
}

// getStackTrace returns a formatted stack trace string.
func getStackTrace() string {
	var sb strings.Builder
	pcs := make([]uintptr, 10)
	n := runtime.Callers(3, pcs) // skip getStackTrace, ErrorLogFields, and caller
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}

// LogInfo logs an info message with LogContext attributes from context.
func LogInfo(ctx context.Context, msg string, extraArgs ...any) {
	args := contextArgsFromCtx(ctx)
	args = append(args, extraArgs...)
	slog.InfoContext(ctx, msg, args...)
}

// LogError logs an error message with LogContext attributes from context.
func LogError(ctx context.Context, msg string, extraArgs ...any) {
	args := contextArgsFromCtx(ctx)
	args = append(args, extraArgs...)
	slog.ErrorContext(ctx, msg, args...)
}

// LogWarn logs a warning message with LogContext attributes from context.
func LogWarn(ctx context.Context, msg string, extraArgs ...any) {
	args := contextArgsFromCtx(ctx)
	args = append(args, extraArgs...)
	slog.WarnContext(ctx, msg, args...)
}

// contextArgsFromCtx extracts LogContext from context and returns slog args.
func contextArgsFromCtx(ctx context.Context) []any {
	lc, ok := Extract(ctx)
	if !ok {
		return nil
	}

	attrs := lc.ToSlogAttrs()
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return args
}
