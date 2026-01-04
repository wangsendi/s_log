package s_log

import (
	"io"
	"log/slog"
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

type colorFormatter struct {
	fn func(io.Writer, *slog.HandlerOptions) slog.Handler
}

func (f *colorFormatter) Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
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
				color, ok := levelColors[lv]
				if !ok {
					if lv < slog.LevelInfo {
						color = fgGray
					} else {
						color = fgRed
					}
				}
				return slog.String(a.Key, color+lv.String()+reset)
			}
		case slog.MessageKey:
			return slog.String(a.Key, fgCyan+a.Value.String()+reset)
		}
		return a
	}
	return f.fn(w, opts)
}

var (
	jsonFmt = &formatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewJSONHandler(w, opts)
	}}
	textFmt = &formatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewTextHandler(w, opts)
	}}
	colorTextFmt = &colorFormatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewTextHandler(w, opts)
	}}
	colorJSONFmt = &colorFormatter{fn: func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewJSONHandler(w, opts)
	}}
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
