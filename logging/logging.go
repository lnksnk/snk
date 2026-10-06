package logging

import (
	"context"
	slog "log/slog"
)

type Logging interface {
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	Log(lvl slog.Level, msg string, args ...any)
	LogContext(ctx context.Context, lvl slog.Level, msg string, args ...any)
}

type logging struct {
	lggr  *slog.Logger
	hndlr slog.Handler
}

// Log implements [Logging].
func (l *logging) Log(lvl slog.Level, msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(context.Background(), lvl, msg, args...)
}

// LogContext implements [Logging].
func (l *logging) LogContext(ctx context.Context, lvl slog.Level, msg string, args ...any) {
	if lggr := l.lggr; lggr != nil {
		lggr.Log(ctx, lvl, msg, args...)
	}
}

// Debug implements [Logging].
func (l *logging) Debug(msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(context.Background(), slog.LevelDebug, msg, args...)
}

// DebugContext implements [Logging].
func (l *logging) DebugContext(ctx context.Context, msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(ctx, slog.LevelDebug, msg, args...)
}

// Error implements [Logging].
func (l *logging) Error(msg string, args ...any) {
	if l == nil {
		return
	}
	l.LogContext(context.Background(), slog.LevelError, msg, args...)
}

// ErrorContext implements [Logging].
func (l *logging) ErrorContext(ctx context.Context, msg string, args ...any) {
	if l == nil {
		return
	}
	l.LogContext(ctx, slog.LevelError, msg, args...)
}

// Info implements [Logging].
func (l *logging) Info(msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(context.Background(), slog.LevelInfo, msg, args...)
}

// InfoContext implements [Logging].
func (l *logging) InfoContext(ctx context.Context, msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn implements [Logging].
func (l *logging) Warn(msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(context.Background(), slog.LevelWarn, msg, args...)
}

// WarnContext implements [Logging].
func (l *logging) WarnContext(ctx context.Context, msg string, args ...any) {
	if l == nil || msg == "" {
		return
	}
	l.LogContext(ctx, slog.LevelWarn, msg, args...)
}

func LogThis(hndlr slog.Handler) Logging {
	return &logging{hndlr: hndlr, lggr: slog.New(hndlr)}
}

func init() {
	slog.Default()
}
