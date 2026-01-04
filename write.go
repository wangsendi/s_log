package s_log

import (
	"io"
	"os"
	"sync"
	"sync/atomic"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Writer interface {
	io.Writer
	io.Closer
}

var stdoutInstance Writer = &stdoutWriter{}

type stdoutWriter struct{}

func (w *stdoutWriter) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (w *stdoutWriter) Close() error                { return nil }

func Stdout() Writer { return stdoutInstance }

type fileWriter struct{ *lumberjack.Logger }

func (w *fileWriter) Close() error { return nil }

type fileOptions struct {
	maxSize, maxBackups, maxAge int
	compress                    bool
}

type FileOption func(*fileOptions)

func WithRotation(maxSize, maxBackups int) FileOption {
	return func(o *fileOptions) { o.maxSize, o.maxBackups = maxSize, maxBackups }
}

func WithMaxAge(maxAge int) FileOption {
	return func(o *fileOptions) { o.maxAge = maxAge }
}

func WithCompress(compress bool) FileOption {
	return func(o *fileOptions) { o.compress = compress }
}

func File(path string, opts ...FileOption) Writer {
	o := &fileOptions{maxSize: 100, maxBackups: 7, maxAge: 30, compress: true}
	for _, opt := range opts {
		opt(o)
	}
	return &fileWriter{
		Logger: &lumberjack.Logger{
			Filename:   path,
			MaxSize:    o.maxSize,
			MaxBackups: o.maxBackups,
			MaxAge:     o.maxAge,
			Compress:   o.compress,
		},
	}
}

type asyncWriter struct {
	w      Writer
	ch     chan []byte
	wg     sync.WaitGroup
	closed int32
}

func (w *asyncWriter) Write(p []byte) (int, error) {
	if atomic.LoadInt32(&w.closed) != 0 {
		return len(p), nil
	}
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case w.ch <- buf:
	default:
	}
	return len(p), nil
}

func (w *asyncWriter) Close() error {
	if !atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		return nil
	}
	close(w.ch)
	w.wg.Wait()
	return w.w.Close()
}

func Async(w Writer, bufferSize int) Writer {
	aw := &asyncWriter{w: w, ch: make(chan []byte, bufferSize)}
	aw.wg.Add(1)
	go func() {
		defer aw.wg.Done()
		for buf := range aw.ch {
			_, _ = aw.w.Write(buf)
		}
	}()
	return aw
}

type multiWriter struct{ writers []Writer }

func (w *multiWriter) Write(p []byte) (int, error) {
	for _, writer := range w.writers {
		_, _ = writer.Write(p)
	}
	return len(p), nil
}

func (w *multiWriter) Close() error {
	var firstErr error
	for _, writer := range w.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func Multi(writers ...Writer) Writer {
	return &multiWriter{writers: writers}
}
