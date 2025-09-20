package s_log

import (
	"log/slog"
)

type options struct {
	level      slog.Level // 日志等级
	color      bool       // 控制台彩色
	path       string     // 文件路径（为空表示不写文件）
	json       bool       // 文件是否用 JSON 格式
	trace      bool       // 是否启用 trace_id
	env        string     // 环境 dev/test/prod
	caller     bool       // 是否启用调用者信息
	timeFormat string     // 时间格式

	// 文件切割
	maxSize    int
	maxAge     int
	maxBackups int
	compress   bool
}

type Option func(*options)

// WithLevel 设置日志级别
func WithLevel(level string) Option {
	return func(o *options) {
		switch level {
		case "debug":
			o.level = slog.LevelDebug
		case "warn", "warning":
			o.level = slog.LevelWarn
		case "error":
			o.level = slog.LevelError
		default:
			o.level = slog.LevelInfo
		}
	}
}

func WithColor() Option { return func(o *options) { o.color = true } }
func WithFile(path string, json bool) Option {
	return func(o *options) { o.path, o.json = path, json }
}
func WithJson(json bool) Option        { return func(o *options) { o.json = json } }
func WithTrace() Option                { return func(o *options) { o.trace = true } }
func WithCaller() Option               { return func(o *options) { o.caller = true } }
func WithEnv(env string) Option        { return func(o *options) { o.env = env } }
func WithTimeFormat(fmt string) Option { return func(o *options) { o.timeFormat = fmt } }
func WithLumberjack(maxSize, maxAge, maxBackups int, compress bool) Option {
	return func(o *options) {
		o.maxSize = maxSize
		o.maxAge = maxAge
		o.maxBackups = maxBackups
		o.compress = compress
	}
}
func WithConsoleOnly() Option {
	return func(o *options) { o.path = "" }
}
