package s_log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	f := JSON()
	if f == nil {
		t.Fatal("JSON() should not return nil")
	}

	buf := &bytes.Buffer{}
	h := f.Format(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	if h == nil {
		t.Fatal("Format() should not return nil")
	}

	logger := slog.New(h)
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message: %s", output)
	}
	if !strings.Contains(output, "key") {
		t.Errorf("output should contain key: %s", output)
	}
}

func TestText(t *testing.T) {
	f := Text()
	if f == nil {
		t.Fatal("Text() should not return nil")
	}

	buf := &bytes.Buffer{}
	h := f.Format(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	if h == nil {
		t.Fatal("Format() should not return nil")
	}

	logger := slog.New(h)
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message: %s", output)
	}
	if !strings.Contains(output, "key") {
		t.Errorf("output should contain key: %s", output)
	}
}

func TestColorText(t *testing.T) {
	f := ColorText()
	if f == nil {
		t.Fatal("ColorText() should not return nil")
	}

	buf := &bytes.Buffer{}
	h := f.Format(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	if h == nil {
		t.Fatal("Format() should not return nil")
	}

	logger := slog.New(h)
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message: %s", output)
	}
}

func TestColorJSON(t *testing.T) {
	f := ColorJSON()
	if f == nil {
		t.Fatal("ColorJSON() should not return nil")
	}

	buf := &bytes.Buffer{}
	h := f.Format(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	if h == nil {
		t.Fatal("Format() should not return nil")
	}

	logger := slog.New(h)
	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message: %s", output)
	}
}

func TestColorText_WithNilOptions(t *testing.T) {
	f := ColorText()
	buf := &bytes.Buffer{}
	h := f.Format(buf, nil)
	if h == nil {
		t.Fatal("Format() should handle nil options")
	}

	logger := slog.New(h)
	logger.Info("test")
}

func TestColorJSON_WithNilOptions(t *testing.T) {
	f := ColorJSON()
	buf := &bytes.Buffer{}
	h := f.Format(buf, nil)
	if h == nil {
		t.Fatal("Format() should handle nil options")
	}

	logger := slog.New(h)
	logger.Info("test")
}

func TestFormatter_Singleton(t *testing.T) {
	f1 := JSON()
	f2 := JSON()
	if f1 != f2 {
		t.Error("JSON() should return singleton")
	}

	f3 := Text()
	f4 := Text()
	if f3 != f4 {
		t.Error("Text() should return singleton")
	}

	f5 := ColorText()
	f6 := ColorText()
	if f5 != f6 {
		t.Error("ColorText() should return singleton")
	}

	f7 := ColorJSON()
	f8 := ColorJSON()
	if f7 != f8 {
		t.Error("ColorJSON() should return singleton")
	}
}
