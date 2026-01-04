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

func (f *formatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler { return f.fn(w, opts) }

type colorTextHandler struct {
	w      io.Writer
	opts   *slog.HandlerOptions
	level  *slog.LevelVar
	groups []string
}

func (h *colorTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.level == nil || level >= h.level.Level()
}

func (h *colorTextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 512)
	buf = append(buf, "time="...)
	buf = r.Time.AppendFormat(buf, "2006-01-02T15:04:05.000Z07:00")
	buf = append(buf, " level="...)
	c := levelColors[r.Level]
	if c == "" {
		if r.Level < slog.LevelInfo {
			c = fgGray
		} else {
			c = fgRed
		}
	}
	buf = append(append(append(buf, c...), r.Level.String()...), reset...)
	if h.opts != nil && h.opts.AddSource && r.PC != 0 {
		if f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next(); f.File != "" {
			file := f.File
			if wd, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(wd, file); err == nil && len(rel) < len(file) {
					file = rel
				}
			}
			buf = append(buf, " source="...)
			buf = append(buf, fgGray...)
			buf = append(buf, '"')
			buf = append(buf, file...)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(f.Line), 10)
			buf = append(buf, '"')
			buf = append(buf, reset...)
		}
	}
	buf = append(buf, " msg="...)
	buf = append(buf, fgCyan...)
	buf = append(buf, r.Message...)
	buf = append(buf, reset...)
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.LevelKey || a.Key == slog.MessageKey || a.Key == slog.TimeKey || a.Key == slog.SourceKey {
			return true
		}
		buf = append(buf, ' ')
		buf = append(buf, a.Key...)
		buf = append(buf, '=')
		h.appendValue(&buf, a.Value)
		buf = append(buf, reset...)
		return true
	})
	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

func (h *colorTextHandler) appendValue(buf *[]byte, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		s := v.String()
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			*buf = append(*buf, fgGray...)
		}
		*buf = append(*buf, s...)
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
		s := fmt.Sprint(v.Any())
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			*buf = append(*buf, fgGray...)
		}
		*buf = append(*buf, s...)
	case slog.KindLogValuer:
		h.appendValue(buf, v.LogValuer().LogValue())
	case slog.KindGroup:
		for _, a := range v.Group() {
			*buf = append(*buf, ' ')
			*buf = append(*buf, a.Key...)
			*buf = append(*buf, '=')
			h.appendValue(buf, a.Value)
		}
	}
}

func (h *colorTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *colorTextHandler) WithGroup(name string) slog.Handler {
	return &colorTextHandler{h.w, h.opts, h.level, append(h.groups, name)}
}

type colorFormatter struct{}

func (f *colorFormatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	var lv *slog.LevelVar
	if opts != nil && opts.Level != nil {
		lv, _ = opts.Level.(*slog.LevelVar)
	}
	return &colorTextHandler{w, opts, lv, nil}
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
				c := levelColors[lv]
				if c == "" {
					if lv < slog.LevelInfo {
						c = fgGray
					} else {
						c = fgRed
					}
				}
				return slog.String(a.Key, c+lv.String()+reset)
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
	reset    = "\x1b[0m"
	fgRed    = "\x1b[31m"
	fgGreen  = "\x1b[32m"
	fgYellow = "\x1b[33m"
	fgCyan   = "\x1b[36m"
	fgGray   = "\x1b[90m"
)

var levelColors = map[slog.Level]string{
	slog.LevelDebug: fgGray,
	slog.LevelInfo:  fgGreen,
	slog.LevelWarn:  fgYellow,
	slog.LevelError: fgRed,
}
