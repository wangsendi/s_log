package s_log

import (
	"context"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

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

// 公共 ReplaceAttr
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.String("time", a.Value.Time().Format("2006-01-02 15:04:05"))
	}
	return a
}

func newH(cfg *options, out io.Writer) slog.Handler {
	switch {
	case cfg.color:
		return tint.NewHandler(out, &tint.Options{
			Level:      cfg.level,
			AddSource:  cfg.caller,
			TimeFormat: cfg.timeFormat,
		})
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
