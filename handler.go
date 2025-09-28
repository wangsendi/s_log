package s_log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 简单的颜色常量
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

// 简单的彩色 Handler
type colorHandler struct {
	w          io.Writer
	level      slog.Level
	timeFormat string
}

func (h *colorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *colorHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf strings.Builder

	// 时间
	buf.WriteString(gray)
	buf.WriteString(r.Time.Format(h.timeFormat))
	buf.WriteString(reset)
	buf.WriteString(" ")

	// 级别
	switch {
	case r.Level < slog.LevelInfo:
		buf.WriteString(purple + "DEBUG" + reset)
	case r.Level < slog.LevelWarn:
		buf.WriteString(green + "INFO" + reset)
	case r.Level < slog.LevelError:
		buf.WriteString(yellow + "WARN" + reset)
	default:
		buf.WriteString(red + "ERROR" + reset)
	}
	buf.WriteString(" ")

	// 调用者信息
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf.WriteString(cyan)
		buf.WriteString(filepath.Base(f.File))
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%d", f.Line))
		buf.WriteString(reset)
		buf.WriteString(" ")
	}

	// 消息
	buf.WriteString(r.Message)

	// 属性
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != "" {
			buf.WriteString(" ")
			buf.WriteString(blue)
			buf.WriteString(a.Key)
			buf.WriteString(reset)
			buf.WriteString("=")
			buf.WriteString(fmt.Sprintf("%v", a.Value.Any()))
		}
		return true
	})

	buf.WriteString("\n")
	_, err := h.w.Write([]byte(buf.String()))
	return err
}

func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *colorHandler) WithGroup(name string) slog.Handler {
	return h
}

type ctxH struct {
	slog.Handler
	withTrace bool
	withEnv   string
}

func (h *ctxH) Handle(ctx context.Context, r slog.Record) error {
	if h.withTrace {
		if tid, ok := ctx.Value(traceKey).(string); ok && tid != "" {
			r.Add("trace_id", tid)
		}
	}
	if h.withEnv != "" {
		r.Add("env", h.withEnv)
	}
	return h.Handler.Handle(ctx, r)
}

func wrapH(base slog.Handler, withTrace bool, env string) slog.Handler {
	return &ctxH{Handler: base, withTrace: withTrace, withEnv: env}
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.String("time", a.Value.Time().Format("2006-01-02 15:04:05"))
	}
	if a.Key == "source" {
		if _, file, line, ok := runtime.Caller(6); ok {
			file = filepath.Base(file)
			return slog.String("source", fmt.Sprintf("%s:%d", file, line))
		}
	}
	return a
}

func newH(cfg *options, out io.Writer) slog.Handler {
	switch {
	case cfg.color:
		return &colorHandler{
			w:          out,
			level:      cfg.level,
			timeFormat: cfg.timeFormat,
		}
	case cfg.json:
		return slog.NewJSONHandler(out, &slog.HandlerOptions{
			Level:       cfg.level,
			AddSource:   cfg.caller,
			ReplaceAttr: replaceAttr,
		})
	default:
		return slog.NewTextHandler(out, &slog.HandlerOptions{
			Level:       cfg.level,
			AddSource:   cfg.caller,
			ReplaceAttr: replaceAttr,
		})
	}
}

func fileW(cfg *options) io.Writer {
	return &lumberjack.Logger{
		Filename:   cfg.path,
		MaxSize:    cfg.maxSize,
		MaxBackups: cfg.maxBackups,
		MaxAge:     cfg.maxAge,
		Compress:   cfg.compress,
	}
}
