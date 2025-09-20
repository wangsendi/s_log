package s_log

import "context"

type ctxKey string

const traceKey ctxKey = "trace_id"

// 为上下文添加追踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey, traceID)
}

// 从上下文中获取追踪ID
func GetTraceID(ctx context.Context) (string, bool) {
	tid, ok := ctx.Value(traceKey).(string)
	return tid, ok
}
