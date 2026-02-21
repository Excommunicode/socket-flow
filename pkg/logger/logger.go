package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/caitlinelfring/go-env-default"
)

type ctxKey string

const (
	TraceIDKey   ctxKey = "trace_id"
	RequestIDKey ctxKey = "request_id"
	TenantIDKey  ctxKey = "tenant_id"
	ThreadKey    ctxKey = "thread"
)

const (
	envLogLevel = "LOG_LEVEL"
)

// parseLogLevel возвращает slog.Level по строке из LOG_LEVEL (DEBUG, INFO, WARN, ERROR).
func parseLogLevel(s string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// prettyHandler выводит логи в читаемом формате: время | уровень | сообщение | key=value...
type prettyHandler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
	// Цвета ANSI (пустые строки, если не TTY)
	colors map[slog.Level]string
	reset  string
}

func newPrettyHandler(w io.Writer, level slog.Level, useColor bool) *prettyHandler {
	h := &prettyHandler{w: w, level: level}
	if useColor {
		h.colors = map[slog.Level]string{
			slog.LevelDebug: "\033[36m", // cyan
			slog.LevelInfo:  "\033[32m", // green
			slog.LevelWarn:  "\033[33m", // yellow
			slog.LevelError: "\033[31m", // red
		}
		h.reset = "\033[0m"
	} else {
		h.colors = make(map[slog.Level]string)
	}
	return h
}

// levelLong возвращает полное имя уровня в верхнем регистре для формата [DEBUG], [INFO], ...
func (h *prettyHandler) levelLong(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARN"
	default:
		return "ERROR"
	}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(ctx context.Context, r slog.Record) error {
	requestID := "-"
	if v, ok := ctx.Value(RequestIDKey).(string); ok && v != "" {
		requestID = v
	}
	tenantID := "-"
	if v, ok := ctx.Value(TenantIDKey).(string); ok && v != "" {
		tenantID = v
	}
	thread := "main"
	if v, ok := ctx.Value(ThreadKey).(string); ok && v != "" {
		thread = v
	}
	class := "-"
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		if f, ok := fs.Next(); ok {
			pkg := filepath.Base(filepath.Dir(f.File))
			class = pkg + ":" + filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
		}
	}

	ts := r.Time.Format("2006-01-02T15:04:05.000")
	lv := r.Level
	levelStr := h.levelLong(lv)
	msg := r.Message

	buf := make([]byte, 0, 384)
	buf = append(buf, '[')
	buf = append(buf, ts...)
	buf = append(buf, "] "...)
	if c, ok := h.colors[lv]; ok {
		buf = append(buf, c...)
	}
	buf = append(buf, "[ "...)
	buf = append(buf, levelStr...)
	buf = append(buf, "] "...)
	if h.reset != "" {
		buf = append(buf, h.reset...)
	}
	buf = append(buf, "[request_id="...)
	buf = append(buf, requestID...)
	buf = append(buf, "] [tenant_id="...)
	buf = append(buf, tenantID...)
	buf = append(buf, "] [thread="...)
	buf = append(buf, thread...)
	buf = append(buf, "] [class="...)
	buf = append(buf, class...)
	buf = append(buf, "] "...)
	buf = append(buf, msg...)

	for _, a := range h.attrs {
		buf = append(buf, "  "...)
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		buf = append(buf, a.Value.String()...)
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.SourceKey {
			return true
		}
		buf = append(buf, "  "...)
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		buf = append(buf, a.Value.String()...)
		return true
	})

	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = append(h2.attrs, attrs...)
	return &h2
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.group = name
	return &h2
}

// isTerminal проверяет, что w — терминал (для цветного вывода).
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isTerminalFile(f)
	}
	return false
}

func isTerminalFile(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

type CustomHandler struct {
	slog.Handler
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

func New() *slog.Logger {
	level := parseLogLevel(env.GetDefault(envLogLevel, "DEBUG"))
	useColor := isTerminal(os.Stdout)
	h := newPrettyHandler(os.Stdout, level, useColor)
	logger := slog.New(&CustomHandler{h})
	return logger
}
