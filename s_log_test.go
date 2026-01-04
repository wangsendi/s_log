package s_log

import (
	"bytes"
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit()
	if globalLogger == nil {
		t.Error("globalLogger should not be nil")
	}
}

func TestMustInit(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit(WithLevel("DEBUG"))
	if globalLogger == nil {
		t.Error("globalLogger should not be nil")
	}
}

func TestClose(t *testing.T) {
	MustInit()
	err := Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestWithLevel(t *testing.T) {
	defer func() { _ = Close() }()

	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"DEBUG", "DEBUG", slog.LevelDebug},
		{"INFO", "INFO", slog.LevelInfo},
		{"WARN", "WARN", slog.LevelWarn},
		{"ERROR", "ERROR", slog.LevelError},
		{"lowercase", "debug", slog.LevelDebug},
		{"invalid", "INVALID", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MustInit(WithLevel(tt.level))
			if levelVar.Level() != tt.expected {
				t.Errorf("expected level %v, got %v", tt.expected, levelVar.Level())
			}
		})
	}
}

func TestWithFormatter(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit(WithFormatter(JSON()))
	if globalLogger == nil {
		t.Error("globalLogger should not be nil")
	}

	MustInit(WithFormatter(ColorText()))
	if globalLogger == nil {
		t.Error("globalLogger should not be nil")
	}
}

func TestWithWriter(t *testing.T) {
	defer func() { _ = Close() }()

	buf := &bytes.Buffer{}
	w := &testWriter{buf: buf}

	MustInit(WithWriter(w))
	slog.Info("test message")

	if buf.Len() == 0 {
		t.Error("writer should receive data")
	}
}

func TestWithAddSource(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit(WithAddSource(true))
	slog.Info("test")

	MustInit(WithAddSource(false))
	slog.Info("test")
}

func TestSetLevel(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit()
	SetLevel("DEBUG")
	if levelVar.Level() != slog.LevelDebug {
		t.Errorf("expected DEBUG, got %v", levelVar.Level())
	}

	SetLevel("ERROR")
	if levelVar.Level() != slog.LevelError {
		t.Errorf("expected ERROR, got %v", levelVar.Level())
	}
}

func TestPresetDev(t *testing.T) {
	defer func() { _ = Close() }()

	opts := PresetDev()
	if len(opts) == 0 {
		t.Error("PresetDev() should return options")
	}

	MustInit(opts...)
	if levelVar.Level() != slog.LevelDebug {
		t.Errorf("expected DEBUG, got %v", levelVar.Level())
	}
}

func TestPresetProd(t *testing.T) {
	defer func() { _ = Close() }()

	opts := PresetProd()
	if len(opts) == 0 {
		t.Error("PresetProd() should return options")
	}

	MustInit(opts...)
	if levelVar.Level() != slog.LevelInfo {
		t.Errorf("expected INFO, got %v", levelVar.Level())
	}
}

func TestPreset(t *testing.T) {
	defer func() { _ = Close() }()

	opts := Preset("WARN", "text", "")
	MustInit(opts...)

	if levelVar.Level() != slog.LevelWarn {
		t.Errorf("expected WARN, got %v", levelVar.Level())
	}
}

func TestPreset_Default(t *testing.T) {
	defer func() { _ = Close() }()

	opts := Preset("", "", "")
	MustInit(opts...)

	if levelVar.Level() != slog.LevelInfo {
		t.Errorf("expected INFO, got %v", levelVar.Level())
	}
}

func TestPreset_WithFile(t *testing.T) {
	defer func() { _ = Close() }()

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.log")
	opts := Preset("DEBUG", "json", file)
	MustInit(opts...)

	if levelVar.Level() != slog.LevelDebug {
		t.Errorf("expected DEBUG, got %v", levelVar.Level())
	}
}

func TestWithInterceptor(t *testing.T) {
	defer func() { _ = Close() }()

	var intercepted bool
	MustInit(WithInterceptor(func(ctx context.Context, r *Record) *Record {
		intercepted = true
		r.Attrs = append(r.Attrs, slog.String("test", "value"))
		return r
	}))

	slog.Info("test message")

	if !intercepted {
		t.Error("interceptor should be called")
	}
}

func TestWithInterceptor_Filter(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit(WithInterceptor(func(ctx context.Context, r *Record) *Record {
		if strings.Contains(r.Message, "filter") {
			return nil
		}
		return r
	}))

	buf := &bytes.Buffer{}
	w := &testWriter{buf: buf}
	MustInit(WithWriter(w), WithInterceptor(func(ctx context.Context, r *Record) *Record {
		if strings.Contains(r.Message, "filter") {
			return nil
		}
		return r
	}))

	slog.Info("filter this")
	slog.Info("keep this")

	if strings.Contains(buf.String(), "filter") {
		t.Error("filtered message should not appear")
	}
	if !strings.Contains(buf.String(), "keep") {
		t.Error("non-filtered message should appear")
	}
}

func TestWithRequestID(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit()
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-request-id")

	if ctx.Value(contextKey{}) != "test-request-id" {
		t.Error("request ID should be set in context")
	}
}

func TestFromContext(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit()
	ctx := context.Background()

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("logger should not be nil")
	}

	ctx = WithRequestID(ctx, "test-id")
	logger = FromContext(ctx)
	if logger == nil {
		t.Error("logger should not be nil")
	}
}

func TestMultipleInit(t *testing.T) {
	defer func() { _ = Close() }()

	MustInit(WithLevel("DEBUG"))
	MustInit(WithLevel("INFO"))

	if levelVar.Level() != slog.LevelInfo {
		t.Errorf("expected INFO, got %v", levelVar.Level())
	}
}

type testWriter struct {
	buf *bytes.Buffer
}

func (w *testWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *testWriter) Close() error {
	return nil
}
