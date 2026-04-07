package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"

	"github.com/illikainen/attest/internal/textutil"
)

type HandlerOptions struct {
	Name      string
	AddSource bool
	Level     slog.Leveler
	NoPrefix  bool
}

type SanitizedHandler struct {
	*HandlerOptions

	writer io.Writer
	attrs  []slog.Attr
}

func NewSanitizedHandler(w io.Writer, opts *HandlerOptions) *SanitizedHandler {
	return &SanitizedHandler{
		HandlerOptions: opts,
		writer:         w,
	}
}

func (h *SanitizedHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.Level.Level() <= level
}

// revive:disable-next-line
func (h *SanitizedHandler) Handle(_ context.Context, record slog.Record) error { //nolint
	prefix := ""
	if !h.NoPrefix {
		if h.Name != "" {
			prefix += "\033[35m" + h.Name + "\033[0m: "
		}

		switch record.Level {
		case slog.LevelDebug:
			prefix += "\033[36mdebug\033[0m: "
		case slog.LevelInfo:
			prefix += "\033[32minfo\033[0m: "
		case slog.LevelWarn:
			prefix += "\033[33mwarn\033[0m: "
		case slog.LevelError:
			prefix += "\033[31merror\033[0m: "
		default:
			prefix += "\033[31minvalid\033[0m: "
		}
	}

	var attrs []string
	for _, attr := range h.attrs {
		attrs = append(attrs, attr.String())
	}

	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, strings.Trim(attr.String(), " \t\r\n"))
		return true
	})

	if h.AddSource {
		frames := runtime.CallersFrames([]uintptr{record.PC})
		frame, _ := frames.Next()
		attrs = append(attrs, "func="+frame.Function)
	}

	msg := prefix + textutil.Sanitize(record.Message)
	if len(attrs) > 0 {
		msg += " | " + textutil.Sanitize(strings.Join(attrs, ", "))
	}
	msg += "\n"

	if n, err := h.writer.Write([]byte(msg)); err != nil || n != len(msg) {
		return fmt.Errorf("bad write (%d): %w", n, err)
	}

	return nil
}

func (h *SanitizedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SanitizedHandler{
		HandlerOptions: h.HandlerOptions,
		writer:         h.writer,
		attrs:          append(attrs, h.attrs...),
	}
}

func (h *SanitizedHandler) WithGroup(_ string) slog.Handler {
	panic("not implemented")
}

func ParseLevel(s string) slog.Level {
	level := new(slog.Level)
	if err := level.UnmarshalText([]byte(strings.ToUpper(s))); err != nil {
		*level = slog.LevelDebug
	}
	return *level
}
