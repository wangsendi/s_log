package s_log

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	// 测试默认配置
	err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 测试带配置的初始化
	err = New(
		WithLevel("debug"),
		WithColor(),
		WithCaller(),
		WithEnv("test"),
	)
	if err != nil {
		t.Fatalf("New() with options failed: %v", err)
	}
}

func TestLogFunctions(t *testing.T) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志
	err := New(WithLevel("debug"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 测试各种日志级别
	Debug("这是一条调试日志")
	Info("这是一条信息日志")
	Warn("这是一条警告日志")
	Error("这是一条错误日志")

	// 测试带参数的日志
	Info("用户登录", "user_id", 123, "ip", "192.168.1.1")
	Error("数据库连接失败", "error", "connection timeout", "retry_count", 3)
}

func TestContextLogFunctions(t *testing.T) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志
	err := New(WithLevel("debug"), WithTrace())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 创建带trace_id的上下文
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-12345")

	// 测试带上下文的日志函数
	DebugContext(ctx, "带上下文的调试日志")
	InfoContext(ctx, "带上下文的信息日志", "user_id", 456)
	WarnContext(ctx, "带上下文的警告日志")
	ErrorContext(ctx, "带上下文的错误日志", "error", "test error")
}

func TestTraceID(t *testing.T) {
	ctx := context.Background()

	// 测试添加trace_id
	traceID := "test-trace-123"
	ctx = WithTraceID(ctx, traceID)

	// 测试获取trace_id
	retrievedTraceID, ok := GetTraceID(ctx)
	if !ok {
		t.Fatal("GetTraceID() should return true")
	}
	if retrievedTraceID != traceID {
		t.Fatalf("GetTraceID() = %s, want %s", retrievedTraceID, traceID)
	}

	// 测试空上下文
	emptyCtx := context.Background()
	_, ok = GetTraceID(emptyCtx)
	if ok {
		t.Fatal("GetTraceID() on empty context should return false")
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected func(*options) bool
	}{
		{
			name:   "WithLevel debug",
			option: WithLevel("debug"),
			expected: func(o *options) bool {
				return o.level.String() == "DEBUG"
			},
		},
		{
			name:   "WithLevel info",
			option: WithLevel("info"),
			expected: func(o *options) bool {
				return o.level.String() == "INFO"
			},
		},
		{
			name:   "WithLevel warn",
			option: WithLevel("warn"),
			expected: func(o *options) bool {
				return o.level.String() == "WARN"
			},
		},
		{
			name:   "WithLevel error",
			option: WithLevel("error"),
			expected: func(o *options) bool {
				return o.level.String() == "ERROR"
			},
		},
		{
			name:   "WithColor",
			option: WithColor(),
			expected: func(o *options) bool {
				return o.color == true
			},
		},
		{
			name:   "WithFile",
			option: WithFile("test.log", true),
			expected: func(o *options) bool {
				return o.path == "test.log" && o.json == true
			},
		},
		{
			name:   "WithTrace",
			option: WithTrace(),
			expected: func(o *options) bool {
				return o.trace == true
			},
		},
		{
			name:   "WithCaller",
			option: WithCaller(),
			expected: func(o *options) bool {
				return o.caller == true
			},
		},
		{
			name:   "WithEnv",
			option: WithEnv("prod"),
			expected: func(o *options) bool {
				return o.env == "prod"
			},
		},
		{
			name:   "WithTimeFormat",
			option: WithTimeFormat("2006-01-02"),
			expected: func(o *options) bool {
				return o.timeFormat == "2006-01-02"
			},
		},
		{
			name:   "WithLumberjack",
			option: WithLumberjack(200, 60, 10, false),
			expected: func(o *options) bool {
				return o.maxSize == 200 && o.maxAge == 60 && o.maxBackups == 10 && o.compress == false
			},
		},
		{
			name:   "WithConsoleOnly",
			option: WithConsoleOnly(),
			expected: func(o *options) bool {
				return o.path == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &options{}
			tt.option(cfg)
			if !tt.expected(cfg) {
				t.Errorf("option %s did not set expected values", tt.name)
			}
		})
	}
}

func TestFileOutput(t *testing.T) {
	// 创建临时文件
	tmpFile := "test_log_output.log"
	defer os.Remove(tmpFile)

	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志，输出到文件
	err := New(
		WithFile(tmpFile, false), // 使用文本格式
		WithLevel("info"),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 写入一些日志
	Info("测试文件输出", "key", "value")
	Error("测试错误日志", "error", "test error")

	// 检查文件是否存在
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("日志文件未创建")
	}

	// 读取文件内容
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "测试文件输出") {
		t.Error("日志文件中未找到预期内容")
	}
	if !strings.Contains(contentStr, "测试错误日志") {
		t.Error("日志文件中未找到预期内容")
	}
}

func TestJSONOutput(t *testing.T) {
	// 创建临时文件
	tmpFile := "test_json_output.log"
	defer os.Remove(tmpFile)

	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志，输出JSON格式到文件
	err := New(
		WithFile(tmpFile, true), // 使用JSON格式
		WithLevel("info"),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 写入一些日志
	Info("测试JSON输出", "key", "value", "number", 123)

	// 读取文件内容
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	contentStr := string(content)
	// 检查是否包含JSON格式的内容
	if !strings.Contains(contentStr, `"msg":"测试JSON输出"`) {
		t.Error("JSON日志格式不正确")
	}
	if !strings.Contains(contentStr, `"key":"value"`) {
		t.Error("JSON日志格式不正确")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 设置日志级别为WARN
	err := New(WithLevel("warn"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 这些日志应该被过滤掉（不会输出）
	Debug("这条调试日志不应该出现")
	Info("这条信息日志不应该出现")

	// 这些日志应该输出
	Warn("这条警告日志应该出现")
	Error("这条错误日志应该出现")
}

func TestConcurrentLogging(t *testing.T) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志
	err := New(WithLevel("debug"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// 并发写入日志
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Info("并发日志测试", "goroutine", id, "iteration", j)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkLogging(b *testing.B) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志
	err := New(WithLevel("info"))
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("基准测试日志", "iteration", i)
	}
}

func BenchmarkContextLogging(b *testing.B) {
	// 重置日志器
	log = nil
	logOnce = sync.Once{}

	// 初始化日志
	err := New(WithLevel("info"), WithTrace())
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}

	ctx := WithTraceID(context.Background(), "bench-trace")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InfoContext(ctx, "基准测试上下文日志", "iteration", i)
	}
}
