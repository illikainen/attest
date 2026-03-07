package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

var levels = map[string]int{
	"trace": LevelTrace,
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

type Options struct {
	Name       string
	Level      string
	level      int
	Writers    []io.Writer
	ShowPrefix bool
}

var opts = Options{
	Name:       "",
	Level:      "trace",
	level:      LevelTrace,
	Writers:    []io.Writer{os.Stderr},
	ShowPrefix: true,
}

func Setup(o *Options) {
	writers := opts.Writers
	opts = *o
	if opts.Writers == nil {
		opts.Writers = writers
	}

	if level, ok := levels[strings.ToLower(opts.Level)]; ok {
		opts.level = level
	} else {
		opts.level = LevelTrace
	}
}

func Tracef(format string, v ...any) {
	printf(LevelTrace, format, v...)
}

func Debugf(format string, v ...any) {
	printf(LevelDebug, format, v...)
}

func Infof(format string, v ...any) {
	printf(LevelInfo, format, v...)
}

func Warnf(format string, v ...any) {
	printf(LevelWarn, format, v...)
}

func Errorf(format string, v ...any) {
	printf(LevelError, format, v...)
}

func printf(level int, format string, v ...any) {
	if level < opts.level {
		return
	}

	prefix := ""
	if opts.ShowPrefix {
		if opts.Name != "" {
			prefix += "\033[35m" + opts.Name + "\033[0m: "
		}

		switch level {
		case LevelTrace:
			prefix += "\033[36mtrace\033[0m: "
		case LevelDebug:
			prefix += "\033[36mdebug\033[0m: "
		case LevelInfo:
			prefix += "\033[32minfo\033[0m: "
		case LevelWarn:
			prefix += "\033[33mwarn\033[0m: "
		default:
			prefix += "\033[31merror\033[0m: "
		}
	}

	for _, line := range strings.Split(strings.TrimRight(fmt.Sprintf(format, v...), "\n"), "\n") {
		msg := fmt.Sprintf("%s%s\n", prefix, sanitize(line))

		for _, w := range opts.Writers {
			if n, err := w.Write([]byte(msg)); err != nil || n != len(msg) {
				panic(err)
			}
		}
	}
}

func sanitize(s string) string {
	var result []byte

	for _, b := range []byte(s) {
		if b != 0x0a && (b < 0x20 || b > 0x7e) {
			result = append(result, '_')
		} else {
			result = append(result, b)
		}
	}

	return string(result)
}

func SetVerbosity(s string) error {
	level, ok := levels[strings.ToLower(s)]
	if !ok {
		var names []string
		for name := range levels {
			names = append(names, name)
		}
		return fmt.Errorf("%s", strings.Join(names, ", "))
	}

	opts.level = level
	return nil
}
