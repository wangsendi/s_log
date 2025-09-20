# s_log

一个基于 Go 标准库 `log/slog` 的增强日志库，提供了更丰富的功能和更友好的使用体验。

## 功能特性

- 🎨 **彩色控制台输出** - 使用 `tint` 库提供美观的彩色日志输出
- 📁 **文件日志支持** - 支持将日志写入文件，支持 JSON 和文本格式
- 🔄 **日志轮转** - 基于 `lumberjack` 实现日志文件自动轮转和压缩
- 🔍 **链路追踪** - 支持 trace_id 自动注入，便于分布式系统调试
- 🌍 **环境标识** - 支持环境标识（dev/test/prod）自动添加
- 📍 **调用者信息** - 可选择是否显示调用者文件和行号
- ⚙️ **灵活配置** - 支持多种配置选项，满足不同场景需求

## 安装

```bash
go get github.com/wangsendi/s_log
```

## 快速开始

### 基础使用

```go
package main

import (
    "context"
    "github.com/wangsendi/s_log"
)

func main() {
    // 初始化日志库
    err := s_log.New()
    if err != nil {
        panic(err)
    }
    
    // 使用默认日志记录器
    s_log.Info("这是一条信息日志")
    s_log.Error("这是一条错误日志")
    s_log.Debug("这是一条调试日志")
}
```

### 高级配置

```go
package main

import (
    "context"
    "github.com/wangsendi/s_log"
)

func main() {
    // 配置日志库
    err := s_log.New(
        s_log.WithLevel("debug"),           // 设置日志级别
        s_log.WithColor(),                  // 启用彩色输出
        s_log.WithFile("logs/app.log", true), // 输出到文件，使用JSON格式
        s_log.WithTrace(),                  // 启用trace_id
        s_log.WithCaller(),                 // 显示调用者信息
        s_log.WithEnv("dev"),               // 设置环境标识
        s_log.WithLumberjack(100, 30, 5, true), // 配置日志轮转
    )
    if err != nil {
        panic(err)
    }
    
    // 使用带上下文的日志
    ctx := context.Background()
    ctx = s_log.WithTraceID(ctx, "trace-12345")
    
    s_log.InfoContext(ctx, "带trace_id的日志")
}
```

## API 文档

### 初始化函数

#### `New(opts ...Option) error`

初始化日志库，支持多种配置选项。

### 配置选项

| 选项 | 函数 | 说明 |
|------|------|------|
| 日志级别 | `WithLevel(level string)` | 设置日志级别：debug/info/warn/error |
| 彩色输出 | `WithColor()` | 启用控制台彩色输出 |
| 文件输出 | `WithFile(path string, json bool)` | 输出到文件，支持JSON格式 |
| JSON格式 | `WithJson(json bool)` | 设置是否使用JSON格式 |
| 链路追踪 | `WithTrace()` | 启用trace_id自动注入 |
| 调用者信息 | `WithCaller()` | 显示调用者文件和行号 |
| 环境标识 | `WithEnv(env string)` | 设置环境标识 |
| 时间格式 | `WithTimeFormat(fmt string)` | 自定义时间格式 |
| 日志轮转 | `WithLumberjack(maxSize, maxAge, maxBackups int, compress bool)` | 配置日志文件轮转 |
| 仅控制台 | `WithConsoleOnly()` | 仅输出到控制台，不写文件 |

### 上下文相关

#### `WithTraceID(ctx context.Context, traceID string) context.Context`

为上下文添加追踪ID。

#### `GetTraceID(ctx context.Context) (string, bool)`

从上下文中获取追踪ID。

## 配置示例

### 开发环境配置

```go
s_log.New(
    s_log.WithLevel("debug"),
    s_log.WithColor(),
    s_log.WithCaller(),
    s_log.WithEnv("dev"),
)
```

### 生产环境配置

```go
s_log.New(
    s_log.WithLevel("info"),
    s_log.WithFile("logs/app.log", true),
    s_log.WithTrace(),
    s_log.WithEnv("prod"),
    s_log.WithLumberjack(100, 30, 5, true),
)
```

### 测试环境配置

```go
s_log.New(
    s_log.WithLevel("warn"),
    s_log.WithFile("logs/test.log", false),
    s_log.WithEnv("test"),
)
```

## 依赖

- `github.com/lmittmann/tint` - 彩色控制台输出
- `gopkg.in/natefinch/lumberjack.v2` - 日志文件轮转

## 许可证

MIT License