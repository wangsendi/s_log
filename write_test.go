package s_log

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestStdout(t *testing.T) {
	w := Stdout()
	if w == nil {
		t.Fatal("Stdout() should not return nil")
	}

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestStdout_Singleton(t *testing.T) {
	w1 := Stdout()
	w2 := Stdout()
	if w1 != w2 {
		t.Error("Stdout() should return singleton")
	}
}

func TestFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	w := File(path)
	if w == nil {
		t.Fatal("File() should not return nil")
	}

	testData := []byte("test log message")
	n, err := w.Write(testData)
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Contains(content, testData) {
		t.Errorf("file should contain test data")
	}
}

func TestFile_WithRotation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	w := File(path, WithRotation(1, 3))
	if w == nil {
		t.Fatal("File() should not return nil")
	}

	err := w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestFile_WithMaxAge(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	w := File(path, WithMaxAge(7))
	if w == nil {
		t.Fatal("File() should not return nil")
	}

	err := w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestFile_WithCompress(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	w := File(path, WithCompress(true))
	if w == nil {
		t.Fatal("File() should not return nil")
	}

	err := w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestFile_MultipleOptions(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	w := File(path,
		WithRotation(10, 5),
		WithMaxAge(30),
		WithCompress(false),
	)
	if w == nil {
		t.Fatal("File() should not return nil")
	}

	err := w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestAsync(t *testing.T) {
	buf := &bytes.Buffer{}
	baseWriter := &testWriter{buf: buf}

	w := Async(baseWriter, 10)
	if w == nil {
		t.Fatal("Async() should not return nil")
	}

	testData := []byte("test message")
	n, err := w.Write(testData)
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	time.Sleep(100 * time.Millisecond)

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), testData) {
		t.Error("async writer should write data")
	}
}

func TestAsync_Concurrent(t *testing.T) {
	buf := &bytes.Buffer{}
	baseWriter := &testWriter{buf: buf}

	w := Async(baseWriter, 100)
	defer func() { _ = w.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := []byte(strings.Repeat("test", id+1))
			_, _ = w.Write(data)
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	if buf.Len() == 0 {
		t.Error("async writer should handle concurrent writes")
	}
}

func TestAsync_CloseMultipleTimes(t *testing.T) {
	buf := &bytes.Buffer{}
	baseWriter := &testWriter{buf: buf}

	w := Async(baseWriter, 10)
	_, _ = w.Write([]byte("test"))

	err1 := w.Close()
	err2 := w.Close()

	if err1 != nil {
		t.Errorf("first Close() failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close() failed: %v", err2)
	}
}

func TestAsync_WriteAfterClose(t *testing.T) {
	buf := &bytes.Buffer{}
	baseWriter := &testWriter{buf: buf}

	w := Async(baseWriter, 10)
	_ = w.Close()

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write() after Close() should not fail: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}
}

func TestMulti(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	w1 := &testWriter{buf: buf1}
	w2 := &testWriter{buf: buf2}

	w := Multi(w1, w2)
	if w == nil {
		t.Fatal("Multi() should not return nil")
	}

	testData := []byte("test message")
	n, err := w.Write(testData)
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("expected %d bytes written, got %d", len(testData), n)
	}

	if !bytes.Contains(buf1.Bytes(), testData) {
		t.Error("first writer should receive data")
	}
	if !bytes.Contains(buf2.Bytes(), testData) {
		t.Error("second writer should receive data")
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestMulti_Empty(t *testing.T) {
	w := Multi()
	if w == nil {
		t.Fatal("Multi() should not return nil")
	}

	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestMulti_MultipleWriters(t *testing.T) {
	buffers := make([]*bytes.Buffer, 5)
	writers := make([]Writer, 5)
	for i := 0; i < 5; i++ {
		buffers[i] = &bytes.Buffer{}
		writers[i] = &testWriter{buf: buffers[i]}
	}

	w := Multi(writers...)
	testData := []byte("test message")
	_, _ = w.Write(testData)

	for i, buf := range buffers {
		if !bytes.Contains(buf.Bytes(), testData) {
			t.Errorf("writer %d should receive data", i)
		}
	}

	_ = w.Close()
}
