package s_log

import (
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
