# s_log

基于 Go 标准库 `log/slog` 的结构化日志库，采用 Functional Options 模式，提供简洁优雅的 API。

## 特性

- **多种格式化器** - JSON、Text、彩色输出
- **灵活的输出目标** - 标准输出、文件、异步写入、多目标
- **日志轮转** - 基于 lumberjack 的自动轮转和压缩
- **拦截器支持** - 自定义日志处理逻辑
- **Context 集成** - 支持请求追踪
- **动态级别** - 运行时调整日志级别
- **并发安全** - 所有操作都是线程安全的
- **极简设计** - 代码简洁，无冗余

## 安装

```bash
go get github.com/wangsendi/s_log
```

## 快速开始

### 最简单的使用

```go
package main

import (
	"log/slog"
	"github.com/wangsendi/s_log"
)

func main() {
	// 使用默认配置（INFO 级别，文本格式，标准输出）
	s_log.MustInit()
	defer s_log.Close()

	slog.Info("应用启动", "version", "1.0.0")
	slog.Debug("这条日志不会输出", "level", "DEBUG")
}
```

### 使用预设配置

```go
package main

import (
	"log/slog"
	"github.com/wangsendi/s_log"
)

func main() {
	// 开发环境：彩色输出 + DEBUG + 源代码位置
	s_log.MustInit(s_log.PresetDev()...)
	defer s_log.Close()

	slog.Info("应用启动", "version", "1.0.0")
	slog.Debug("调试信息", "key", "value")
	slog.Warn("警告信息", "warning", "something")
	slog.Error("错误信息", "error", "something wrong")
}
```

### 自定义配置

```go
s_log.MustInit(
	s_log.WithLevel("DEBUG"),
	s_log.WithFormatter(s_log.ColorText()),
	s_log.WithWriter(s_log.Multi(
		s_log.Stdout(),
		s_log.File("app.log", s_log.WithRotation(100, 7)),
	)),
	s_log.WithAddSource(true),
)
```

### 使用 Preset 配置

`Preset` 函数接受三个参数：level、format、file，空字符串使用默认值：

```go
s_log.MustInit(s_log.Preset("INFO", "json", "/var/log/app.log")...)
```

参数说明：

- `level`: 日志级别 (DEBUG/INFO/WARN/ERROR)，默认 INFO
- `format`: 格式 (json/text/color)，默认 json
- `file`: 日志文件路径，空字符串则输出到标准输出

示例：

```go
// 使用默认值（INFO, json, stdout）
s_log.MustInit(s_log.Preset("", "", "")...)

// 自定义配置
s_log.MustInit(s_log.Preset("DEBUG", "color", "")...)

// 输出到文件
s_log.MustInit(s_log.Preset("INFO", "json", "/var/log/app.log")...)
```

## 详细文档

### 初始化

#### `MustInit(opts ...Option)`

初始化日志系统，失败时 panic。适合在应用启动时使用。

```go
s_log.MustInit(s_log.PresetDev()...)
```

#### `Close() error`

关闭日志系统，释放资源。建议使用 `defer` 确保资源被正确释放。

```go
defer s_log.Close()
```

### 配置选项

| 函数                                       | 说明                                 |
| ------------------------------------------ | ------------------------------------ |
| `WithLevel(level string)`                  | 设置日志级别 (DEBUG/INFO/WARN/ERROR) |
| `WithFormatter(f Formatter)`               | 设置格式化器                         |
| `WithWriter(w Writer)`                     | 设置输出目标                         |
| `WithAddSource(on bool)`                   | 是否显示源代码位置                   |
| `WithInterceptor(interceptor Interceptor)` | 设置拦截器                           |

### 格式化器

| 函数          | 说明                     |
| ------------- | ------------------------ |
| `JSON()`      | JSON 格式，适合生产环境  |
| `Text()`      | 键值对格式，兼容传统工具 |
| `ColorText()` | 彩色文本，适合开发环境   |
| `ColorJSON()` | 彩色 JSON，适合终端调试  |

#### JSON 格式示例

```go
s_log.MustInit(
	s_log.WithFormatter(s_log.JSON()),
	s_log.WithWriter(s_log.Stdout()),
)

slog.Info("用户登录", "user_id", 123, "ip", "192.168.1.1")
// 输出: {"time":"2024-01-01T10:00:00Z","level":"INFO","msg":"用户登录","user_id":123,"ip":"192.168.1.1"}
```

#### Text 格式示例

```go
s_log.MustInit(
	s_log.WithFormatter(s_log.Text()),
	s_log.WithWriter(s_log.Stdout()),
)

slog.Info("用户登录", "user_id", 123, "ip", "192.168.1.1")
// 输出: time=2024-01-01T10:00:00Z level=INFO msg="用户登录" user_id=123 ip=192.168.1.1
```

#### ColorText 格式示例

```go
s_log.MustInit(
	s_log.WithFormatter(s_log.ColorText()),
	s_log.WithWriter(s_log.Stdout()),
)

slog.Info("用户登录", "user_id", 123)
// 输出带颜色的文本，INFO 为绿色，消息为青色
```

#### ColorJSON 格式示例

```go
s_log.MustInit(
	s_log.WithFormatter(s_log.ColorJSON()),
	s_log.WithWriter(s_log.Stdout()),
)

slog.Info("用户登录", "user_id", 123)
// 输出带颜色的 JSON，适合终端调试
```

### 输出目标

| 函数                                    | 说明               |
| --------------------------------------- | ------------------ |
| `Stdout()`                              | 标准输出           |
| `File(path string, opts ...FileOption)` | 文件输出，支持轮转 |
| `Async(w Writer, bufferSize int)`       | 异步写入           |
| `Multi(writers ...Writer)`              | 多目标输出         |

#### File 选项

| 函数                                    | 说明                       | 默认值 |
| --------------------------------------- | -------------------------- | ------ |
| `WithRotation(maxSize, maxBackups int)` | 设置轮转参数（MB, 备份数） | 100, 7 |
| `WithMaxAge(maxAge int)`                | 设置保留天数               | 30     |
| `WithCompress(compress bool)`           | 是否压缩旧日志             | true   |

#### 文件输出示例

```go
s_log.MustInit(
	s_log.WithWriter(s_log.File(
		"app.log",
		s_log.WithRotation(100, 7),    // 100MB 轮转，保留 7 个备份
		s_log.WithMaxAge(30),          // 保留 30 天
		s_log.WithCompress(true),       // 压缩旧日志
	)),
)
```

#### 异步写入示例

异步写入可以提高性能，适合高并发场景。当缓冲区满时，新日志会被丢弃（非阻塞）：

```go
s_log.MustInit(
	s_log.WithWriter(s_log.Async(
		s_log.File("app.log", s_log.WithRotation(100, 7)),
		1000, // 缓冲区大小
	)),
)
```

#### 多目标输出示例

同时输出到标准输出和文件。`Multi` 可以接受任意数量的 Writer，包括空参数：

```go
s_log.MustInit(
	s_log.WithWriter(s_log.Multi(
		s_log.Stdout(),
		s_log.File("app.log", s_log.WithRotation(100, 7)),
	)),
)
```

### 预设配置

| 函数                          | 说明                                    |
| ----------------------------- | --------------------------------------- |
| `PresetDev()`                 | 开发环境：DEBUG + 彩色文本 + 源代码位置 |
| `PresetProd()`                | 生产环境：INFO + JSON + 文件输出        |
| `Preset(level, format, file)` | 自定义配置，参数为空时使用默认值        |

### 动态级别

运行时动态调整日志级别，无需重启应用：

```go
s_log.SetLevel("DEBUG")  // 开启调试日志
s_log.SetLevel("INFO")   // 恢复正常级别
s_log.SetLevel("ERROR")  // 只显示错误
```

### 拦截器

拦截器可以在日志记录前修改或过滤日志，非常适合添加通用字段或实现日志过滤：

#### 添加 trace_id

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		if traceID := getTraceID(ctx); traceID != "" {
			r.Attrs = append(r.Attrs, slog.String("trace_id", traceID))
		}
		return r
	}),
)
```

#### 添加环境信息

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		r.Attrs = append(r.Attrs,
			slog.String("env", "production"),
			slog.String("service", "user-service"),
		)
		return r
	}),
)
```

#### 过滤敏感信息

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		// 过滤掉包含敏感信息的日志
		if strings.Contains(r.Message, "password") {
			return nil // 返回 nil 过滤该日志
		}
		return r
	}),
)
```

#### 修改日志级别

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		// 将某些错误降级为警告
		if r.Level == slog.LevelError && strings.Contains(r.Message, "retry") {
			r.Level = slog.LevelWarn
		}
		return r
	}),
)
```

### Context 集成

在 HTTP 请求等场景中使用，实现请求追踪：

#### HTTP 中间件示例

```go
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := s_log.WithRequestID(r.Context(), requestID)
		log := s_log.FromContext(ctx)

		log.Info("请求开始",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		next.ServeHTTP(w, r.WithContext(ctx))

		log.Info("请求完成",
			"method", r.Method,
			"path", r.URL.Path,
		)
	})
}
```

#### gRPC 拦截器示例

```go
func UnaryLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestID := getRequestIDFromMetadata(ctx)
	ctx = s_log.WithRequestID(ctx, requestID)
	log := s_log.FromContext(ctx)

	log.Info("gRPC 请求",
		"method", info.FullMethod,
	)

	resp, err := handler(ctx, req)

	if err != nil {
		log.Error("gRPC 请求失败",
			"method", info.FullMethod,
			"error", err,
		)
	} else {
		log.Info("gRPC 请求成功",
			"method", info.FullMethod,
		)
	}

	return resp, err
}
```

## 完整示例

### 开发环境配置

```go
package main

import (
	"log/slog"
	"github.com/wangsendi/s_log"
)

func main() {
	s_log.MustInit(s_log.PresetDev()...)
	defer s_log.Close()

	slog.Info("应用启动", "version", "1.0.0")
	slog.Debug("调试信息", "key", "value")
}
```

### 生产环境配置

```go
package main

import (
	"log/slog"
	"github.com/wangsendi/s_log"
)

func main() {
	s_log.MustInit(s_log.PresetProd()...)
	defer s_log.Close()

	slog.Info("应用启动", "version", "1.0.0")
}
```

### 完整自定义配置

```go
package main

import (
	"context"
	"log/slog"
	"github.com/wangsendi/s_log"
)

func main() {
	s_log.MustInit(
		s_log.WithLevel("INFO"),
		s_log.WithFormatter(s_log.JSON()),
		s_log.WithWriter(s_log.Async(
			s_log.File("app.log",
				s_log.WithRotation(100, 7),
				s_log.WithMaxAge(30),
				s_log.WithCompress(true),
			),
			1000,
		)),
		s_log.WithAddSource(false),
		s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
			r.Attrs = append(r.Attrs, slog.String("service", "my-service"))
			return r
		}),
	)
	defer s_log.Close()

	slog.Info("应用启动", "version", "1.0.0")
}
```

### 多目标输出

```go
s_log.MustInit(
	s_log.WithWriter(s_log.Multi(
		s_log.Stdout(),
		s_log.File("app.log", s_log.WithRotation(100, 7)),
		s_log.File("error.log", s_log.WithRotation(50, 3)),
	)),
)
```

### 仅文件输出（无标准输出）

```go
s_log.MustInit(
	s_log.WithWriter(s_log.File("app.log", s_log.WithRotation(100, 7))),
)
```

## 最佳实践

### 1. 应用启动时初始化

```go
func main() {
	s_log.MustInit(s_log.Preset("INFO", "json", "/var/log/app.log")...)
	defer s_log.Close()

	// 其他初始化代码...
}
```

### 2. 使用 defer 确保资源释放

```go
defer s_log.Close()
```

### 3. 生产环境使用 JSON 格式

JSON 格式便于日志收集和分析：

```go
s_log.MustInit(
	s_log.WithFormatter(s_log.JSON()),
	s_log.WithWriter(s_log.File("app.log", s_log.WithRotation(100, 7))),
)
```

### 4. 高并发场景使用异步写入

```go
s_log.MustInit(
	s_log.WithWriter(s_log.Async(
		s_log.File("app.log", s_log.WithRotation(100, 7)),
		1000, // 根据实际情况调整缓冲区大小
	)),
)
```

### 5. 使用拦截器添加通用字段

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		r.Attrs = append(r.Attrs,
			slog.String("service", "user-service"),
			slog.String("version", "1.0.0"),
			slog.String("env", os.Getenv("ENV")),
		)
		return r
	}),
)
```

### 6. 使用 Context 传递请求 ID

```go
ctx := s_log.WithRequestID(ctx, requestID)
log := s_log.FromContext(ctx)
log.Info("处理请求", "path", path)
```

### 7. 文件 Writer 默认配置

`File` Writer 默认配置为：100MB 轮转、保留 7 个备份、保留 30 天、启用压缩。可根据需要覆盖：

```go
s_log.File("app.log") // 使用默认配置
s_log.File("app.log", s_log.WithRotation(50, 3)) // 自定义轮转
```

## 性能说明

- **单例格式化器**: 格式化器使用单例模式，避免重复创建
- **异步写入**: 使用 `Async` Writer 可以显著提高高并发场景下的性能，缓冲区满时丢弃日志（非阻塞）
- **动态级别**: 使用 `slog.LevelVar` 实现无锁的动态级别调整
- **并发安全**: 所有操作都是线程安全的，可以在多个 goroutine 中安全使用
- **极简实现**: 代码简洁高效，无冗余逻辑

## 常见问题

### Q: 如何同时输出到控制台和文件？

A: 使用 `Multi` Writer：

```go
s_log.MustInit(
	s_log.WithWriter(s_log.Multi(
		s_log.Stdout(),
		s_log.File("app.log", s_log.WithRotation(100, 7)),
	)),
)
```

### Q: 如何过滤某些日志？

A: 使用拦截器返回 `nil`：

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		if strings.Contains(r.Message, "debug") {
			return nil // 过滤该日志
		}
		return r
	}),
)
```

### Q: 如何动态调整日志级别？

A: 使用 `SetLevel`：

```go
s_log.SetLevel("DEBUG")  // 开启调试
s_log.SetLevel("INFO")   // 恢复正常
```

### Q: 如何添加自定义字段到所有日志？

A: 使用拦截器：

```go
s_log.MustInit(
	s_log.WithInterceptor(func(ctx context.Context, r *s_log.Record) *s_log.Record {
		r.Attrs = append(r.Attrs, slog.String("custom_field", "value"))
		return r
	}),
)
```

## 架构说明

s_log 采用 Handler + Formatter + Writer 三层架构：

- **Handler**: 基于 `slog.Handler`，处理日志记录逻辑，支持拦截器
- **Formatter**: 决定日志输出格式（JSON/Text/Color），复用标准库实现
- **Writer**: 决定日志输出目标（Stdout/File/Async/Multi），支持组合使用

这种架构设计使得各个组件可以独立使用和组合，提供了极大的灵活性。

## API 参考

### Writer 接口

```go
type Writer interface {
	io.Writer
	io.Closer
}
```

所有 Writer 都实现了 `io.Writer` 和 `io.Closer` 接口。

### Formatter 接口

```go
type Formatter interface {
	Format(w io.Writer, opts *slog.HandlerOptions) slog.Handler
}
```

### Interceptor 类型

```go
type Interceptor func(ctx context.Context, r *Record) *Record
```

返回 `nil` 可以过滤日志，返回修改后的 `Record` 可以修改日志内容。

### Record 类型

```go
type Record struct {
	Time    time.Time
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}
```

## 依赖

- `gopkg.in/natefinch/lumberjack.v2` - 日志文件轮转

## 许可证

MIT License
