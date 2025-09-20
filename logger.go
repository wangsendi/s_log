package s_log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	log     *slog.Logger
	logOnce sync.Once
)

func New(opts ...Option) error {
	var err error
	logOnce.Do(func() {
		cfg := &options{
			level:      slog.LevelInfo,
			timeFormat: "2006-01-02 15:04:05",
			maxSize:    100,
			maxBackups: 5,
			maxAge:     30,
			compress:   true,
		}
		for _, opt := range opts {
			opt(cfg)
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout)
		if cfg.path != "" {
			writers = append(writers, fileW(cfg))
		}
		out := io.MultiWriter(writers...)

		handler := newH(cfg, out)

		log = slog.New(wrapH(handler, cfg.trace, cfg.env))
		slog.SetDefault(log)
	})
	return err
}

// 日志函数
func Debug(msg string, args ...any) {
	if log != nil {
		log.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if log != nil {
		log.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if log != nil {
		log.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if log != nil {
		log.Error(msg, args...)
	}
}

// 带上下文的日志函数
func DebugContext(ctx context.Context, msg string, args ...any) {
	if log != nil {
		log.DebugContext(ctx, msg, args...)
	}
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	if log != nil {
		log.InfoContext(ctx, msg, args...)
	}
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	if log != nil {
		log.WarnContext(ctx, msg, args...)
	}
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	if log != nil {
		log.ErrorContext(ctx, msg, args...)
	}
}
