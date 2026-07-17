package lib

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Custom slog level between Info and Warn (go-logger Notice).
const levelNotice = slog.Level(2)

// Log is the process-wide logger. Output is message-only with ANSI colors by
// level, matching the previous go-logger CLI behaviour.
var Log *AppLogger

func init() {
	Log = newAppLogger(os.Stdout, slog.LevelDebug)
}

// AppLogger wraps slog with the Info/Notice/Warning/Error helpers used across the codebase.
type AppLogger struct {
	logger *slog.Logger
}

func newAppLogger(w io.Writer, level slog.Level) *AppLogger {
	h := &messageHandler{w: w, level: level}
	return &AppLogger{logger: slog.New(h)}
}

func (l *AppLogger) DebugF(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *AppLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *AppLogger) InfoF(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *AppLogger) NoticeF(format string, args ...any) {
	l.logger.Log(context.Background(), levelNotice, fmt.Sprintf(format, args...))
}

func (l *AppLogger) WarningF(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *AppLogger) Warningf(format string, args ...any) {
	l.WarningF(format, args...)
}

func (l *AppLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *AppLogger) ErrorF(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *AppLogger) Errorf(format string, args ...any) {
	l.ErrorF(format, args...)
}

// ANSI colors matching apsdehal/go-logger.
const (
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[37m"
	ansiReset  = "\033[0m"
)

func colorFor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return ansiRed
	case level >= slog.LevelWarn:
		return ansiYellow
	case level >= levelNotice:
		return ansiGreen
	case level >= slog.LevelInfo:
		return ansiWhite
	default:
		return ansiCyan // Debug
	}
}

// messageHandler prints only the log message, with level colors.
type messageHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *messageHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *messageHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := fmt.Fprintf(h.w, "%s%s%s\n", colorFor(r.Level), r.Message, ansiReset)
	return err
}

func (h *messageHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *messageHandler) WithGroup(_ string) slog.Handler { return h }
