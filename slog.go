package s_log

import (
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type HandlerType int

const (
	HandlerText HandlerType = iota
	HandlerJSON
)

type Option func(*config)

type config struct {
	level     slog.Level
	handler   HandlerType
	addSource bool
	logFile   string
	color     bool
}

func defaults() *config {
	return &config{
		level:     slog.LevelInfo,
		handler:   HandlerText,
		addSource: true,
		color:     false,
	}
}

func WithLevel(l slog.Level) Option    { return func(c *config) { c.level = l } }
func WithHandler(h HandlerType) Option { return func(c *config) { c.handler = h } }
func WithAddSource(on bool) Option     { return func(c *config) { c.addSource = on } }
func WithFile(path string) Option      { return func(c *config) { c.logFile = path } }
func WithColor(on bool) Option         { return func(c *config) { c.color = on } }

/* ========= 颜色（仅 Text）========= */

const (
	reset    = "\x1b[0m"
	fgRed    = "\x1b[31m"
	fgGreen  = "\x1b[32m"
	fgYellow = "\x1b[33m"
	fgCyan   = "\x1b[36m"
	fgGray   = "\x1b[90m"
)

func colorize(s, c string) string { return c + s + reset }

func colorReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.LevelKey:
		if lv, ok := a.Value.Any().(slog.Level); ok {
			switch {
			case lv <= slog.LevelDebug:
				return slog.String(a.Key, colorize(lv.String(), fgGray))
			case lv == slog.LevelInfo:
				return slog.String(a.Key, colorize(lv.String(), fgGreen))
			case lv == slog.LevelWarn:
				return slog.String(a.Key, colorize(lv.String(), fgYellow))
			default:
				return slog.String(a.Key, colorize(lv.String(), fgRed))
			}
		}
	case slog.MessageKey:
		return slog.String(a.Key, colorize(a.Value.String(), fgCyan))
	}
	return a
}

var levelVar slog.LevelVar

func newRotateWriter(path string) *lumberjack.Logger {
	if path == "" {
		return nil
	}
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100, // MB
		MaxBackups: 7,
		MaxAge:     30, // days
		Compress:   true,
	}
}

func New(opts ...Option) *slog.Logger {
	cfg := defaults()
	for _, opt := range opts {
		opt(cfg)
	}

	levelVar.Set(cfg.level)
	hopts := &slog.HandlerOptions{
		Level:     &levelVar,
		AddSource: cfg.addSource,
	}
	if cfg.handler == HandlerText && cfg.color {
		hopts.ReplaceAttr = colorReplaceAttr
	}

	var h slog.Handler
	if w := newRotateWriter(cfg.logFile); w != nil {
		if cfg.handler == HandlerJSON {
			h = slog.NewJSONHandler(w, hopts)
		} else {
			h = slog.NewTextHandler(w, hopts)
		}
	} else {
		if cfg.handler == HandlerJSON {
			h = slog.NewJSONHandler(os.Stdout, hopts)
		} else {
			h = slog.NewTextHandler(os.Stdout, hopts)
		}
	}

	l := slog.New(h)
	slog.SetDefault(l)
	return l
}

func SetLevel(l slog.Level) { levelVar.Set(l) }
