package s_log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type Formatter interface {
	Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler
}

type formatter struct {
	fn func(io.Writer, *slog.HandlerOptions) slog.Handler
}

func (f *formatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return f.fn(w, opts)
}

type colorTextHandler struct {
	w         io.Writer
	opts      *slog.HandlerOptions
	level     *slog.LevelVar
	groups    []string
	workDir   string
	workDirOK bool
}

func (h *colorTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.level == nil || level >= h.level.Level()
}

func (h *colorTextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 512)
	buf = append(append(append(buf, "time="...), r.Time.AppendFormat(nil, timeFormat)...), ' ')
	buf = append(append(append(append(append(buf, "level="...), bold...), getLevelColor(r.Level)...), r.Level.String()...), reset...)
	buf = append(buf, ' ')
	h.writeColored(&buf, fgCyan, r.Message)
	r.Attrs(func(a slog.Attr) bool {
		if !builtinKeys[a.Key] {
			buf = append(buf, ' ')
			h.writeColored(&buf, fgBlue, a.Key)
			buf = append(buf, '=')
			h.appendValue(&buf, a.Value, fgCyan)
		}
		return true
	})
	if h.opts != nil && h.opts.AddSource && r.PC != 0 {
		if f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next(); f.File != "" {
			buf = append(buf, ' ')
			h.writeColored(&buf, fgGray, "source=")
			h.writeColored(&buf, fgGray, strconv.Quote(h.formatSourcePath(f.File)+":"+strconv.Itoa(f.Line)))
		}
	}
	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

func (h *colorTextHandler) writeColored(buf *[]byte, color, text string) {
	if color != "" {
		*buf = append(append(append(*buf, color...), text...), reset...)
	} else {
		*buf = append(*buf, text...)
	}
}

func (h *colorTextHandler) formatSourcePath(file string) string {
	if !h.workDirOK {
		if wd, err := os.Getwd(); err == nil {
			h.workDir, h.workDirOK = wd, true
		}
	}
	if h.workDirOK {
		if rel, err := filepath.Rel(h.workDir, file); err == nil && len(rel) < len(file) {
			return rel
		}
	}
	return file
}

func (h *colorTextHandler) appendValue(buf *[]byte, v slog.Value, color string) {
	if color != "" {
		*buf = append(*buf, color...)
	}
	switch v = v.Resolve(); v.Kind() {
	case slog.KindString:
		*buf = append(*buf, v.String()...)
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		*buf = append(*buf, v.Duration().String()...)
	case slog.KindTime:
		*buf = v.Time().AppendFormat(*buf, time.RFC3339Nano)
	case slog.KindAny:
		*buf = append(*buf, fmt.Sprint(v.Any())...)
	case slog.KindLogValuer:
		h.appendValue(buf, v.LogValuer().LogValue(), color)
	case slog.KindGroup:
		for _, a := range v.Group() {
			*buf = append(append(append(*buf, ' '), a.Key...), '=')
			h.appendValue(buf, a.Value, color)
		}
	}
	if color != "" {
		*buf = append(*buf, reset...)
	}
}

func (h *colorTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *colorTextHandler) WithGroup(name string) slog.Handler {
	return &colorTextHandler{w: h.w, opts: h.opts, level: h.level, groups: append(h.groups, name), workDir: h.workDir, workDirOK: h.workDirOK}
}

type colorFormatter struct{}

func (f *colorFormatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	var lv *slog.LevelVar
	if opts != nil && opts.Level != nil {
		lv, _ = opts.Level.(*slog.LevelVar)
	}
	wd, err := os.Getwd()
	return &colorTextHandler{w: w, opts: opts, level: lv, workDir: wd, workDirOK: err == nil}
}

type colorJSONFormatter struct {
	fn func(io.Writer, *slog.HandlerOptions) slog.Handler
}

func (f *colorJSONFormatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	old := opts.ReplaceAttr
	opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if old != nil {
			a = old(groups, a)
		}
		switch a.Key {
		case slog.LevelKey:
			if lv, ok := a.Value.Any().(slog.Level); ok {
				color := getLevelColor(lv)
				return slog.String(a.Key, bold+color+lv.String()+reset)
			}
		case slog.MessageKey:
			return slog.String(a.Key, fgCyan+a.Value.String()+reset)
		}
		return a
	}
	return f.fn(w, opts)
}

var (
	jsonFmt      = &formatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler { return slog.NewJSONHandler(w, opts) }}
	textFmt      = &formatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler { return slog.NewTextHandler(w, opts) }}
	colorTextFmt = &colorFormatter{}
	colorJSONFmt = &colorJSONFormatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler { return slog.NewJSONHandler(w, opts) }}
)

func JSON() Formatter      { return jsonFmt }
func Text() Formatter      { return textFmt }
func ColorText() Formatter { return colorTextFmt }
func ColorJSON() Formatter { return colorJSONFmt }

const (
	reset      = "\x1b[0m"
	bold       = "\x1b[1m"
	fgRed      = "\x1b[31m"
	fgGreen    = "\x1b[32m"
	fgYellow   = "\x1b[33m"
	fgBlue     = "\x1b[34m"
	fgCyan     = "\x1b[36m"
	fgGray     = "\x1b[90m"
	timeFormat = "2006-01-02T15:04:05.000Z07:00"
)

var (
	levelColors = map[slog.Level]string{
		slog.LevelDebug: fgGray,
		slog.LevelInfo:  fgGreen,
		slog.LevelWarn:  fgYellow,
		slog.LevelError: fgRed,
	}
	builtinKeys = map[string]bool{
		slog.LevelKey:   true,
		slog.MessageKey: true,
		slog.TimeKey:    true,
		slog.SourceKey:  true,
	}
)

func getLevelColor(level slog.Level) string {
	if c, ok := levelColors[level]; ok {
		return c
	}
	if level < slog.LevelInfo {
		return fgGray
	}
	return fgRed
}
