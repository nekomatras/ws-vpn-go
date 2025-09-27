package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

func NewLogger(target io.Writer, level slog.Level) *slog.Logger {
	return slog.New(newCustomHandler(level, target))
}

func GetLoggerWithName(baseLogger *slog.Logger, name string) *slog.Logger {
	return baseLogger.With("module", name)
}

type customHandler struct {
	level slog.Level
	out   io.Writer
	attrs []slog.Attr
}

func newCustomHandler(level slog.Level, out io.Writer) *customHandler {
	return &customHandler{level: level, out: out}
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {

	newAttrs := append([]slog.Attr(nil), h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &customHandler{
		level: h.level,
		out:   h.out,
		attrs: newAttrs,
	}
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006/01/02 15:04:05")
	level := levelShort(r.Level)

	module := ""
	for _, a := range h.attrs {
		if a.Key == "module" {
			module = fmt.Sprintf("%v", a.Value)
			break
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "module" && module == "" {
			module = fmt.Sprintf("%v", a.Value)
			return false
		}
		return true
	})

	if module != "" {
		fmt.Fprintf(h.out, "%s [%s] [%s] %s\n", timestamp, level, module, r.Message)
	} else {
		fmt.Fprintf(h.out, "%s [%s] [%s] %s\n", timestamp, level, "---", r.Message)
	}
	return nil
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool { return level >= h.level }
func (h *customHandler) WithGroup(name string) slog.Handler               { return h }

func levelShort(l slog.Level) string {
	switch {
	case l < slog.LevelDebug:
		return "D"
	case l < slog.LevelInfo:
		return "I"
	case l < slog.LevelWarn:
		return "W"
	case l < slog.LevelError:
		return "E"
	default:
		return "E"
	}
}
