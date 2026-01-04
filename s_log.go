package s_log

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var (
	globalLogger *slog.Logger
	globalWriter Writer
	levelVar     slog.LevelVar
	mu           sync.RWMutex
)

type Interceptor func(ctx context.Context, r *Record) *Record

type Record struct {
	Time    time.Time
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}

type Option func(*config)

type config struct {
	level       slog.Level
	fmt         Formatter
	w           Writer
	addSource   bool
	interceptor Interceptor
}

type contextKey struct{}

type handlerWrapper struct {
	slog.Handler
	interceptor Interceptor
}

func (h *handlerWrapper) Handle(ctx context.Context, r slog.Record) error {
	if h.interceptor == nil {
		return h.Handler.Handle(ctx, r)
	}
	rec := &Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		Attrs:   make([]slog.Attr, 0, r.NumAttrs()),
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.Attrs = append(rec.Attrs, a)
		return true
	})
	if rec = h.interceptor(ctx, rec); rec == nil {
		return nil
	}
	nr := slog.NewRecord(rec.Time, rec.Level, rec.Message, 0)
	for _, a := range rec.Attrs {
		nr.AddAttrs(a)
	}
	return h.Handler.Handle(ctx, nr)
}

func WithLevel(level string) Option {
	return func(c *config) { c.level = parseLevel(level) }
}

func WithFormatter(f Formatter) Option {
	return func(c *config) { c.fmt = f }
}

func WithWriter(w Writer) Option {
	return func(c *config) { c.w = w }
}

func WithAddSource(on bool) Option {
	return func(c *config) { c.addSource = on }
}

func WithInterceptor(interceptor Interceptor) Option {
	return func(c *config) { c.interceptor = interceptor }
}

var levelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

func parseLevel(s string) slog.Level {
	if lv, ok := levelMap[strings.ToUpper(s)]; ok {
		return lv
	}
	return slog.LevelInfo
}

func MustInit(opts ...Option) {
	mu.Lock()
	defer mu.Unlock()

	if globalWriter != nil {
		_ = globalWriter.Close()
	}

	cfg := &config{
		level:     slog.LevelInfo,
		fmt:       Text(),
		w:         Stdout(),
		addSource: false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	levelVar.Set(cfg.level)
	h := cfg.fmt.Format(cfg.w, &slog.HandlerOptions{
		Level:     &levelVar,
		AddSource: cfg.addSource,
	})

	if cfg.interceptor != nil {
		h = &handlerWrapper{Handler: h, interceptor: cfg.interceptor}
	}

	globalLogger = slog.New(h)
	slog.SetDefault(globalLogger)
	globalWriter = cfg.w
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if globalWriter != nil {
		return globalWriter.Close()
	}
	return nil
}

func SetLevel(level string) {
	levelVar.Set(parseLevel(level))
}

func PresetDev() []Option {
	return []Option{
		WithLevel("DEBUG"),
		WithFormatter(ColorText()),
		WithWriter(Stdout()),
		WithAddSource(true),
	}
}

func PresetProd() []Option {
	return []Option{
		WithLevel("INFO"),
		WithFormatter(JSON()),
		WithWriter(File("/var/log/app.log", WithRotation(100, 7))),
		WithAddSource(false),
	}
}

func Preset(level, format, file string) []Option {
	if level == "" {
		level = "INFO"
	}
	if format == "" {
		format = "json"
	}

	opts := []Option{WithLevel(level)}

	switch strings.ToLower(format) {
	case "text":
		opts = append(opts, WithFormatter(Text()))
	case "color":
		opts = append(opts, WithFormatter(ColorText()))
	default:
		opts = append(opts, WithFormatter(JSON()))
	}

	if file != "" {
		opts = append(opts, WithWriter(File(file, WithRotation(100, 7))))
	} else {
		opts = append(opts, WithWriter(Stdout()))
	}
	return opts
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestID)
}

func FromContext(ctx context.Context) *slog.Logger {
	if requestID, ok := ctx.Value(contextKey{}).(string); ok {
		return globalLogger.With("request_id", requestID)
	}
	return globalLogger
}
